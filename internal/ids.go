package internal

import (
	"context"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const consumerIds = "consumer:ids"

type Ids struct {
	rdb *redis.Client
}

func NewIds(rdb *redis.Client) *Ids {
	return &Ids{rdb}
}

func (ids *Ids) Next(ctx context.Context) (string, error) {
	id := uuid.New().String()
	err := ids.rdb.RPush(ctx, consumerIds, id).Err()
	return id, err
}

func (ids *Ids) Clear(ctx context.Context) {
	<-ctx.Done()
	err := ids.rdb.Del(ctx, consumerIds).Err()
	if err != nil {
		ERROR.Printf("Error deleting redis list %v : %v\n", consumerIds, err)
	} else {
		INFO.Printf("Cleaned up redis list %v\n", consumerIds)
	}
}
