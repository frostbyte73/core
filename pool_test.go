package core

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	p := NewQueuePool(2, QueueWorkerParams{QueueSize: 10})

	const (
		key1 = "1"
		key2 = "2"
	)

	var val1, val2 uint32

	val := func(ind ...int) uint32 {
		var v uint32
		for _, i := range ind {
			v += uint32(1 << i)
		}
		return v
	}
	add1 := func(i int) {
		atomic.AddUint32(&val1, 1<<i)
	}
	add2 := func(i int) {
		atomic.AddUint32(&val2, 1<<i)
	}
	get1 := func() uint32 {
		return atomic.LoadUint32(&val1)
	}
	get2 := func() uint32 {
		return atomic.LoadUint32(&val2)
	}

	submitted := p.Submit(key1, func() {
		time.Sleep(time.Millisecond * 500)
		add1(1)
	})
	require.True(t, submitted)

	submitted = p.Submit(key2, func() {
		add2(2)
	})
	require.True(t, submitted)

	submitted = p.Submit(key1, func() {
		add1(3)
	})
	require.True(t, submitted)

	time.Sleep(time.Millisecond * 100)
	require.Equal(t, val(), get1())
	require.Equal(t, val(2), get2())

	time.Sleep(time.Millisecond * 500)
	require.Equal(t, val(1, 3), get1())

	submitted = p.Submit(key1, func() {
		time.Sleep(time.Millisecond * 500)
		add1(4)
	})
	require.True(t, submitted)

	submitted = p.Submit(key2, func() {
		time.Sleep(time.Millisecond * 500)
		add2(5)
	})
	require.True(t, submitted)

	p.Drain()
	require.Equal(t, val(1, 3, 4), get1())
	require.Equal(t, val(2, 5), get2())
}

func TestPoolOverflow(t *testing.T) {
	queueSize := 2
	p := NewQueuePool(2, QueueWorkerParams{QueueSize: queueSize, DropWhenFull: true})

	for i := 0; i < queueSize+2; i++ {
		submitted := p.Submit("key", func() {
			time.Sleep(time.Millisecond * 500)
		})
		if i < 3 {
			require.True(t, submitted)
		} else {
			require.False(t, submitted)
		}
	}

	p.Drain()
}
