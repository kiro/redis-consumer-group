package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
)

func ProcessMessage(ctx context.Context, rdb *redis.Client, consumerId string, msg *redis.Message) error {
	msgJson := make(map[string]string)
	err := json.Unmarshal([]byte(msg.Payload), &msgJson)
	if err != nil {
		return fmt.Errorf("unable to parse json %v : %v\n", msg.Payload, err)
	}
	msgJson["consumer_id"] = consumerId

	err = rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "messages:processed",
		ID:     "*",
		Values: msgJson,
	}).Err()
	if err != nil {
		return fmt.Errorf("unable to add message to messages:processed : %v", err)
	}
	return nil
}
