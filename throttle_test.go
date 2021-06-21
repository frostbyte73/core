package throttle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestThrottle(t *testing.T) {
	throttler := New(time.Millisecond * 100)
	var c int

	// executes immediately
	throttler(func() { c += 17 })
	assert.Equal(t, 17, c)

	time.Sleep(time.Millisecond * 50)
	// queued - 50ms remaining
	throttler(func() { c += 1 })
	assert.Equal(t, 17, c)

	time.Sleep(time.Millisecond * 20)
	// queued - 30ms remaining (overwrites function 2)
	throttler(func() { c += 5 })
	assert.Equal(t, 17, c)

	time.Sleep(time.Millisecond * 30)
	// function 3 executes
	assert.Equal(t, 22, c)

	time.Sleep(time.Millisecond * 110)
	// ready

	// executes immediately
	throttler(func() { c += 4 })
	assert.Equal(t, 26, c)
}

func TestAsync(t *testing.T) {
	throttler := New(time.Millisecond * 100)
	var c int

	// executes immediately
	throttler(func() {
		go func() {
			time.Sleep(time.Millisecond * 150)
			c += 17
		}()
	})

	time.Sleep(time.Millisecond * 50)
	// queued - 50ms remaining
	throttler(func() {
		go func() {
			time.Sleep(time.Millisecond * 100)
			c += 1
		}()
	})

	time.Sleep(time.Millisecond * 20)
	// queued - 30ms remaining (overwrites function 2)
	throttler(func() {
		go func() {
			time.Sleep(time.Millisecond * 50)
			c += 5
		}()
	})

	time.Sleep(time.Millisecond * 30)
	// function 3 executes

	time.Sleep(time.Millisecond * 110)
	// ready

	// executes immediately
	throttler(func() {
		go func() {
			time.Sleep(time.Millisecond * 150)
			c += 4
		}()
	})

	// wait for goroutines
	time.Sleep(time.Millisecond * 150)
	assert.Equal(t, 26, c)
}
