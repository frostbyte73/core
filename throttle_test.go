package core

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestThrottle(t *testing.T) {
	throttler := NewThrottle(time.Millisecond * 100)
	var c uint32
	add := func(v int) {
		atomic.AddUint32(&c, uint32(v))
	}
	get := func() int {
		return int(atomic.LoadUint32(&c))
	}

	// executes immediately
	throttler(func() { add(17) })
	require.Equal(t, 17, get())

	time.Sleep(time.Millisecond * 50)
	// queued - 50ms remaining
	throttler(func() { add(1) })
	require.Equal(t, 17, get())

	time.Sleep(time.Millisecond * 20)
	// queued - 30ms remaining (overwrites function 2)
	throttler(func() { add(5) })
	require.Equal(t, 17, get())

	time.Sleep(time.Millisecond * 40)
	// function 3 executes
	require.Equal(t, 22, get())

	time.Sleep(time.Millisecond * 100)
	// ready

	// executes immediately
	throttler(func() { add(4) })
	require.Equal(t, 26, get())
}

func TestThrottleAsync(t *testing.T) {
	throttler := NewThrottle(time.Millisecond * 100)
	var c uint32
	add := func(v int) {
		atomic.AddUint32(&c, uint32(v))
	}
	get := func() int {
		return int(atomic.LoadUint32(&c))
	}

	// executes immediately
	throttler(func() {
		go func() {
			time.Sleep(time.Millisecond * 150)
			add(17)
		}()
	})

	time.Sleep(time.Millisecond * 50)
	// queued - 50ms remaining
	throttler(func() {
		go func() {
			time.Sleep(time.Millisecond * 100)
			add(1)
		}()
	})

	time.Sleep(time.Millisecond * 20)
	// overwrites function 2 - 30ms remaining
	throttler(func() {
		go func() {
			time.Sleep(time.Millisecond * 50)
			add(5)
		}()
	})

	time.Sleep(time.Millisecond * 140)
	// function 3 executes, throttle ready again
	throttler(func() {
		go func() {
			time.Sleep(time.Millisecond * 150)
			add(4)
		}()
	})

	// wait for function 4
	time.Sleep(time.Millisecond * 160)
	require.Equal(t, 26, get())
}
