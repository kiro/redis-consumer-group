package internal

import (
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const consumerIds = "consumer:ids"

// Ids - Maintains the list of ids of the running consumers.
type Ids struct {
	rdb *redis.Client
}

// NewIds - constructs new ids generator that maintains list of ids redis
func NewIds(rdb *redis.Client) *Ids {
	return &Ids{rdb}
}

// Next - returns the next id and updates the consumer ids in redis
func (ids *Ids) Next(ctx context.Context) (string, error) {
	id := uuid.New().String()
	err := ids.rdb.RPush(ctx, consumerIds, id).Err()
	return id, err
}

// Clear - clears the ids list in redis
func (ids *Ids) Clear() error {
	return ids.rdb.Del(context.Background(), consumerIds).Err()
}
