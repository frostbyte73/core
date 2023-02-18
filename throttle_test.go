package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestThrottle(t *testing.T) {
	throttler := NewThrottle(time.Millisecond * 100)
	var c int

	// executes immediately
	throttler(func() { c += 17 })
	require.Equal(t, 17, c)

	time.Sleep(time.Millisecond * 50)
	// queued - 50ms remaining
	throttler(func() { c += 1 })
	require.Equal(t, 17, c)

	time.Sleep(time.Millisecond * 20)
	// queued - 30ms remaining (overwrites function 2)
	throttler(func() { c += 5 })
	require.Equal(t, 17, c)

	time.Sleep(time.Millisecond * 30)
	// function 3 executes
	require.Equal(t, 22, c)

	time.Sleep(time.Millisecond * 110)
	// ready

	// executes immediately
	throttler(func() { c += 4 })
	require.Equal(t, 26, c)
}

func TestThrottleAsync(t *testing.T) {
	throttler := NewThrottle(time.Millisecond * 100)
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
	// overwrites function 2 - 30ms remaining
	throttler(func() {
		go func() {
			time.Sleep(time.Millisecond * 50)
			c += 5
		}()
	})

	time.Sleep(time.Millisecond * 140)
	// function 3 executes, throttle ready again
	throttler(func() {
		go func() {
			time.Sleep(time.Millisecond * 150)
			c += 4
		}()
	})

	// wait for function 4
	time.Sleep(time.Millisecond * 160)
	require.Equal(t, 26, c)
}
