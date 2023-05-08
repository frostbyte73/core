package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPool(t *testing.T) {
	p := NewQueuePool(QueueWorkerParams{QueueSize: 10})

	key1 := "1"
	key2 := "2"

	res1 := ""
	res2 := ""

	p.Submit(key1, func() {
		time.Sleep(time.Millisecond * 500)
		res1 += "1"
	})

	p.Submit(key2, func() {
		res2 += "2"
	})

	p.Submit(key1, func() {
		res1 += "3"
	})

	time.Sleep(time.Millisecond * 100)
	require.Equal(t, "", res1)
	require.Equal(t, "2", res2)

	time.Sleep(time.Millisecond * 500)
	require.Equal(t, "13", res1)

	p.Submit(key1, func() {
		time.Sleep(time.Millisecond * 500)
		res1 += "4"
	})

	p.Submit(key2, func() {
		time.Sleep(time.Millisecond * 500)
		res2 += "5"
	})

	p.Drain()
	require.Equal(t, "134", res1)
	require.Equal(t, "25", res2)
}
