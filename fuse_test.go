package core

import (
	"testing"
	"time"
)

func TestFuse(t *testing.T) {
	runOnce := func(f *Fuse) <-chan struct{} {
		onceStarted := make(chan struct{})
		onceFinished := make(chan struct{})
		if f.IsBroken() {
			t.Fail()
		}
		go func() {
			f.Once(func() {
				close(onceStarted)
				time.Sleep(time.Second / 2)
				close(onceFinished)
			})
		}()
		<-onceStarted
		if f.IsBroken() {
			t.Fail()
		}
		return onceFinished
	}
	t.Run("break waits for once", func(t *testing.T) {
		var f Fuse
		onceFinished := runOnce(&f)

		f.Break()
		if !f.IsBroken() {
			t.Fail()
		}
		select {
		case <-onceFinished:
		default:
			t.Fatal("break doest wait for once to complete")
		}
	})
	t.Run("can get channel early", func(t *testing.T) {
		var f Fuse
		done := f.Watch()

		<-runOnce(&f)
		if !f.IsBroken() {
			t.Fail()
		}
		select {
		case <-done:
		default:
			t.Fatal("watched channel not closed after once")
		}
	})
}
