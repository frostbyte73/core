package core

import (
	"testing"
	"time"
)

func TestFuse(t *testing.T) {
	var f Fuse

	onceStarted := make(chan struct{})
	onceFinished := make(chan struct{})
	go func() {
		f.Once(func() {
			close(onceStarted)
			time.Sleep(time.Second)
			close(onceFinished)
		})
	}()

	<-onceStarted
	f.Break()
	select {
	case <-onceFinished:
	default:
		t.Fail()
	}
}
