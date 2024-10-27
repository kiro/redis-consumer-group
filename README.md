# Redis consumer group

## /internal

Just put everything in internal as there is no need to reuse the code for this exercise

- ids.go - maintains the list of consumer ids in consumer:ids list in redis. When requesting new id it adds it to the list
and has a method to clear the list.

- counter.go - thread safe counter over the last second. It maintains a bucket with count for each millisecond and total
count and expires the buckets and updates the total as time progresses. Complexity O(1) for all operations, sometimes
bigger constant depending on how often the counter is updated.

- process_message.go - function to read message from pubsub channel, add consumer id field and publish it to processed
stream

- consumer_group.go - subscribes to pubsub "messages:published", runs n goroutines reading from the channel of the
subscription (which guarantees each message will be processed once). Requests the ids for each goroutine and keeps them
in redis and returns a function to clear the state. The consumer group takes a process message function so it can be
reusable if it needs to. It updates the counter as it reads messages and prints it every 3 second. When the passed
context is cancelled everything stops.

- consumer_group_test.go and ids_test.go has functional test using miniredis and mocking out time, logs, etc.
There is a unit test for counter_test.go . To run tests in the root folder of the project
``` 
go test ./...
 ```

To test if increasing the consumers increases the throughput I changed the publisher to send 80K msg in a batch and
run a bit longer. The results look like

- 1 consumer - up to 18000 msgs/s
- 2 consumers - up to 30000 msgs/s
- 4 consumers - up to 50000 msgs/s
- 8 consumers - up to 65000 msgs/s

## /cmd/redis-consumer-group

Executable for the consumer group, has flags --redis-addr by default localhost:6379 and --consumers by default 3 . It
catches when the binary is interrupted with ctrl+C and clears the consumerIds in redis.

To get the project and run the binary

```
git clone https://github.com/kiro/redis-consumer-group
cd redis-consumer-group
go build ./cmd/redis-consumer-group
./redis-consumer-group --consumers=4
```

## Task 4
1. Tests are provided
2. It looks like we can run redis cluster. Have sharded pub/sub for the published messages with topics like
messages:published:0, messages:published:1 ... .  The publisher can do a round robin on the topics and publish on each 
or we can have a publisher per topic. We can have a consumer group instance running for each sharded pub/sub topic and 
publish in stream like messages:processed:0 ... . 
3. Consumer groups already exist for redis streams, so that can be used instead.