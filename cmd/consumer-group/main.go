package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"redis/internal"
)

func main() {
	redisAddr := flag.String("redis-addr", "localhost:6379", "Address of redis instance")
	consumers := flag.Int("consumers", 3, "Number of consumers")
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	cleanState := internal.RunConsumerGroup(ctx, *consumers, *redisAddr)

	// clean the customer ids on control C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
		cleanState()
		os.Exit(1)
	}()
}
