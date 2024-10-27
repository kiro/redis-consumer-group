package internal

import (
	"bytes"
	"context"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"log"
	"slices"
	"strings"
	"testing"
	"time"
)

var (
	stdout     *bytes.Buffer
	stderr     *bytes.Buffer
	rdb        *redis.Client
	ticker     chan time.Time
	unixMilli  int64
	cleanState func() error
	uuids      = []string{
		"61616161-6161-4161-a161-616161616161",
		"62626262-6262-4262-a262-626262626262",
		"62626262-6262-4262-a363-636363636363",
	}
	lastId = "$"
)

func publishAndAssertProcessed(t *testing.T, id string) {
	err := rdb.Publish(context.Background(), "messages:published", fmt.Sprintf("{\"message_id\":\"%v\"}", id)).Err()
	assert.Nil(t, err)

	streams, err := rdb.XRead(context.Background(), &redis.XReadArgs{
		Streams: []string{"messages:processed"},
		Count:   1,
		Block:   1 * time.Second,
		ID:      lastId,
	}).Result()
	lastId = id
	assert.Nil(t, err)
	values := streams[0].Messages[0].Values
	assert.Equal(t, id, values["message_id"])
	assert.True(t, slices.Contains(uuids, values["consumer_id"].(string)))
}

func assertConsumerIds(t *testing.T, uuids []string) {
	n, err := rdb.LLen(context.Background(), "consumer:ids").Result()
	assert.Nil(t, err)
	consumerIds, err := rdb.LRange(context.Background(), "consumer:ids", 0, n).Result()
	assert.Nil(t, err)
	assert.Equal(t, uuids, consumerIds)
}

func TestRunConsumerGroup(t *testing.T) {
	assertConsumerIds(t, uuids)
	ticker <- time.Now()
	publishAndAssertProcessed(t, "1")
	publishAndAssertProcessed(t, "2")
	ticker <- time.Now()
	unixMilli += 1
	publishAndAssertProcessed(t, "3")
	ticker <- time.Now()
	unixMilli += 990
	publishAndAssertProcessed(t, "4")
	ticker <- time.Now()
	unixMilli += 10
	ticker <- time.Now()

	assertConsumerIds(t, uuids)
	cleanState()
	assertConsumerIds(t, []string{})

	assert.Equal(t,
		`INFO Starting consumer with id 61616161-6161-4161-a161-616161616161.
INFO Starting consumer with id 62626262-6262-4262-a262-626262626262.
INFO Starting consumer with id 62626262-6262-4262-a363-636363636363.
INFO Processed 0 messages in the last second.
INFO Processed 2 messages in the last second.
INFO Processed 3 messages in the last second.
INFO Processed 2 messages in the last second.
INFO Processed 2 messages in the last second.
`, stdout.String())

	assert.Equal(t, "", stderr.String())
}

func TestMain(m *testing.M) {
	// setting logs to point to strings for the test
	defer func(info *log.Logger, error *log.Logger) {
		logInfo = info
		logError = error
	}(logInfo, logError)
	stdout = new(bytes.Buffer)
	logInfo = log.New(stdout, "INFO ", log.Lmsgprefix)
	stderr = new(bytes.Buffer)
	logError = log.New(stderr, "ERROR ", log.Lmsgprefix)

	// setting random reader to constant value, so it produces the same uuids
	rep := strings.Repeat
	uuid.SetRand(strings.NewReader(rep("a", 16) + rep("b", 24) + rep("c", 64)))
	defer func() { uuid.SetRand(nil) }()

	// test redis instance
	testRedis, err := miniredis.Run()
	if err != nil {
		log.Panicln(err)
	}
	defer testRedis.Close()
	rdb = redis.NewClient(&redis.Options{
		Addr: testRedis.Addr(),
	})

	// setup fake clock
	ticker = make(chan time.Time)
	unixMilli = time.Now().UnixMilli()
	clock := &clock{
		unixMillis: func() int64 { return unixMilli },
		newTicker:  func(_ time.Duration) <-chan time.Time { return ticker },
	}

	// run the consumer group
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cleanState = runConsumerGroup(ctx, 3, testRedis.Addr(), clock, ProcessMessage)

	m.Run()
}
