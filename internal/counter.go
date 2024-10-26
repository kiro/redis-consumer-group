package internal

import (
	"sync"
	"time"
)

const millisPerSecond = 1000

type bucket struct {
	unixMilli int64
	count     int
}

type Counter struct {
	buckets []bucket
	total   int
	lock    *sync.Mutex
}

// NewCounter - creates counter over the last second.
//
// It maintains buckets with count for every millisecond in the last second.
// The buckets older than a second are removed on every operation and
// a total counter is maintained.
// The number of buckets is selected so the counter is precise up to a millisecond
// and it's not too expensive to do the operations up to O(1000) .
func NewCounter() *Counter {
	return &Counter{[]bucket{}, 0, &sync.Mutex{}}
}

func (c *Counter) clear(unixMilli int64) {
	index := 0
	for i, bucket := range c.buckets {
		if unixMilli-bucket.unixMilli > millisPerSecond {
			index = i + 1
			c.total -= bucket.count
		}
	}
	c.buckets = c.buckets[:index]
}

// Increment - increments the counter
func (c *Counter) Increment() {
	c.lock.Lock()
	defer c.lock.Unlock()

	unixMilli := time.Now().UnixMilli()
	c.clear(unixMilli)
	numBuckets := len(c.buckets)
	if numBuckets > 0 && c.buckets[numBuckets].unixMilli == unixMilli {
		c.buckets[numBuckets].count++
	} else {
		c.buckets = append(c.buckets, bucket{unixMilli, 1})
	}
	c.total++
}

// Get - returns the value of the counter for the last second
func (c *Counter) Get() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.clear(time.Now().UnixMilli())
	return c.total
}
