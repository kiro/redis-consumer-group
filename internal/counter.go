package internal

import (
	"sync"
)

const millisPerSecond = 1000

type bucket struct {
	unixMilli int64
	count     int
}

type Counter struct {
	buckets   []bucket
	total     int
	lock      *sync.Mutex
	unixMilli func() int64
}

// NewCounter - creates counter over the last second.
//
// It maintains buckets with count for every millisecond in the last second.
// The buckets older than a second are removed on every operation and
// a total counter is maintained.
// The number of buckets is selected so the counter is precise up to a millisecond,
// and it's not too expensive to do the operations up to O(1000) .
func NewCounter(unixMilli func() int64) *Counter {
	return &Counter{[]bucket{}, 0, &sync.Mutex{}, unixMilli}
}

func (c *Counter) clear(unixMilli int64) {
	index := 0
	for i, bucket := range c.buckets {
		if unixMilli-bucket.unixMilli > millisPerSecond {
			index = i + 1
			c.total -= bucket.count
		} else {
			break
		}
	}
	c.buckets = c.buckets[index:]
}

// Increment - increments the counter
func (c *Counter) Increment() {
	c.lock.Lock()
	defer c.lock.Unlock()

	unixMilli := c.unixMilli()
	c.clear(unixMilli)
	numBuckets := len(c.buckets)
	if numBuckets > 0 && c.buckets[numBuckets-1].unixMilli == unixMilli {
		c.buckets[numBuckets-1].count++
	} else {
		c.buckets = append(c.buckets, bucket{unixMilli, 1})
	}
	c.total++
}

// Get - returns the value of the counter for the last second
func (c *Counter) Get() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.clear(c.unixMilli())
	return c.total
}
