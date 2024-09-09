package core

import (
	"sync"
	"time"
)

// Throttle is a function throttler that takes a function as its argument.
// If ready, it will execute immediately, and it will always wait the specified duration
// between executions. If multiple functions are added within the same execution window,
// only the last function added will be executed.
type Throttle func(f func())

func NewThrottle(period time.Duration) Throttle {
	t := &throttle{
		period: period,
		ready:  true,
	}

	return func(f func()) {
		t.add(f)
	}
}

type throttle struct {
	m      sync.Mutex
	period time.Duration
	ready  bool
	timer  *time.Timer
	next   func()
}

func (t *throttle) add(f func()) {
	t.m.Lock()
	ready := t.ready
	if ready {
		t.ready = false
		t.timer = time.AfterFunc(t.period, t.execute)
	} else {
		t.next = f
	}
	t.m.Unlock()

	if ready {
		f()
	}
}

func (t *throttle) execute() {
	t.m.Lock()
	f := t.next
	if f != nil {
		t.next = nil
		t.timer = time.AfterFunc(t.period, t.execute)
	} else {
		t.ready = true
	}
	t.m.Unlock()

	if f != nil {
		f()
	}
}
