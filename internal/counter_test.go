package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var nowUnixMilli = int64(time.Now().UnixMilli())

// Adding test with fake time to avoid time flakiness.
func TestCounter(t *testing.T) {
	counter := NewCounter(func() int64 { return nowUnixMilli })

	assert.Equal(t, 0, counter.Get())

	counter.Increment()
	counter.Increment()
	assert.Equal(t, 2, counter.Get())

	nowUnixMilli += 10
	counter.Increment()
	assert.Equal(t, 3, counter.Get())

	nowUnixMilli += 990
	counter.Increment()
	assert.Equal(t, 4, counter.Get())

	nowUnixMilli += 1
	assert.Equal(t, 2, counter.Get())

	nowUnixMilli += 10
	assert.Equal(t, 1, counter.Get())
}

func TestWithRealTime(t *testing.T) {
	counter := NewCounter(func() int64 { return time.Now().UnixMilli() })

	assert.Equal(t, 0, counter.Get())
	counter.Increment()
	assert.Equal(t, 1, counter.Get())
	time.Sleep(1 * time.Millisecond)
	counter.Increment()
	assert.Equal(t, 2, counter.Get())
}
