package internal

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"time"
)

var (
	logInfo  = log.New(os.Stdout, "INFO ", log.Llongfile|log.Ldate|log.Ltime)
	logError = log.New(os.Stderr, "ERROR ", log.Llongfile|log.Ldate|log.Ltime)
)

// clock interface with fake time
type clock struct {
	unixMillis func() int64
	newTicker  func(duration time.Duration) <-chan time.Time
}

func printCounterEvery(ctx context.Context, counter *Counter, duration time.Duration, clock *clock) {
	ticker := clock.newTicker(duration)
	for {
		select {
		case <-ticker:
			logInfo.Printf("Processed %v messages in the last second.\n", counter.Get())
		case <-ctx.Done():
			break
		}
	}
}

func consume(ctx context.Context, rdb *redis.Client, id string, messages <-chan *redis.Message, counter *Counter) {
	logInfo.Printf("Starting consumer with id %v.\n", id)

	for msg := range messages {
		msgJson := make(map[string]string)
		err := json.Unmarshal([]byte(msg.Payload), &msgJson)
		if err != nil {
			logError.Printf("Unable to parse json %v : %v\n", msg.Payload, err)
			continue
		}
		msgJson["consumer_id"] = id

		err = rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: "messages:processed",
			ID:     msgJson["message_id"],
			Values: msgJson,
		}).Err()
		if err != nil {
			logError.Printf("Failed to add processed message : %v", err)
			continue
		} else {
			counter.Increment()
		}
	}
}

func runConsumerGroup(ctx context.Context, n int, redisAddr string, clock *clock) func() error {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	counter := NewCounter(clock.unixMillis)
	go printCounterEvery(ctx, counter, 3*time.Second, clock)

	ids := NewIds(rdb)

	messages := rdb.Subscribe(ctx, "messages:published").Channel()
	for i := 0; i < n; i++ {
		id, err := ids.Next(ctx)
		if err != nil {
			logError.Printf("Unable to get id for consumer %v : %v\n", i, err)
			continue
		}

		go consume(ctx, rdb, id, messages, counter)
	}

	return ids.Clear
}

// RunConsumerGroup - Runs a consumer group of n consumers.
// Returns a function that has to be called to clean the state of the consumer group when the program is terminated.
func RunConsumerGroup(ctx context.Context, n int, redisAddr string) func() error {
	return runConsumerGroup(ctx, n, redisAddr, &clock{
		unixMillis: func() int64 { return time.Now().UnixMilli() },
		newTicker: func(duration time.Duration) <-chan time.Time {
			return time.NewTicker(duration).C
		}})
}
