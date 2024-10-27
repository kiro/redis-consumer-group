package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"redis/internal"
)

func main() {
	redisAddr := flag.String("redis-addr", "localhost:6379", "Address of redis instance")
	consumers := flag.Int("consumers", 3, "Number of consumers")
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	cleanConsumerGroupState := internal.RunConsumerGroup(ctx, *consumers, *redisAddr, internal.ProcessMessage)

	// clean the customer ids when the program gets terminated
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	cancel()
	err := cleanConsumerGroupState()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(1)
}
