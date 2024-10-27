package internal

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"testing"
)

func assertList(t *testing.T, rdb *redis.Client, values ...string) {
	list, err := rdb.LRange(context.Background(), consumerIds, 0, 1).Result()
	assert.Nil(t, err)
	if values == nil {
		values = []string{}
	}
	assert.Equal(t, values, list)
}

func TestIds(t *testing.T) {
	testRedis := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: testRedis.Addr(),
	})
	ids := NewIds(rdb)
	ctx := context.Background()

	// add one id
	id1, err := ids.Next(ctx)
	assert.Nil(t, err)
	assertList(t, rdb, id1)

	// add second id
	id2, err := ids.Next(ctx)
	assert.Nil(t, err)
	assertList(t, rdb, id1, id2)

	// clear
	err = ids.Clear()
	assert.Nil(t, err)
	assertList(t, rdb)
}
