package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func printCounterEvery(ctx context.Context, counter *Counter, duration time.Duration) {
	timer := time.NewTimer(duration)

	for {
		select {
		case <-timer.C:
			fmt.Printf("Processed %v messages", counter.Get())
		case <-ctx.Done():
			timer.Stop()
			break
		}
	}
}

func consumer(ctx context.Context, rdb *redis.Client, id string, messages <-chan *redis.Message) {
	INFO.Printf("Starting consumer with id %v\n", id)

	for {
		msg, ok := <-messages
		if !ok {
			break
		}
		msgJson := make(map[string]string)
		err := json.Unmarshal([]byte(msg.Payload), &msgJson)
		if err != nil {
			ERROR.Printf("Error unmarshalling json %v : %v\n", err, msg.Payload)
			continue
		}
		msgJson["consumer_id"] = id

		err = rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: "messages:processed",
			ID:     msgJson["message_id"],
			Values: msgJson,
		}).Err()
		if err != nil {
			ERROR.Printf("Error adding processed message : %v", err)
			continue
		}
	}
}

// Runs a consumer group of n consumers.
func RunConsumerGroup(ctx context.Context, n int, redisAddr string) func() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	counter := NewCounter()
	go printCounterEvery(ctx, counter, 3*time.Second)

	ids := NewIds(rdb)

	messages := rdb.Subscribe(ctx, "messages:published").Channel()
	for i := 0; i < n; i++ {
		id, err := ids.Next(ctx)
		if err != nil {
			ERROR.Printf("Unable to get id for consumer %v : %v\n", i, err)
			continue
		}

		go consumer(ctx, rdb, id, messages)
	}

	return func() { ids.Clear(ctx) }
}
