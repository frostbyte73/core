package throttle

import (
	"sync"
	"time"
)

// New returns a function throttler that takes a function as its argument.
// If ready, it will execute immediately, and it will always wait the specified duration
// between executions. The last function added will be the function executed.
func New(after time.Duration) func(f func()) {
	t := &throttler{
		after: after,
		ready: true,
	}

	return func(f func()) {
		t.add(f)
	}
}

type throttler struct {
	mu    sync.Mutex
	after time.Duration
	ready bool
	timer *time.Timer
	next  func()
}

func (t *throttler) add(f func()) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.ready {
		t.ready = false
		f()
		t.timer = time.AfterFunc(t.after, t.execute)
	} else {
		t.next = f
	}
}

func (t *throttler) execute() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.next != nil {
		t.next()
		t.next = nil
		t.timer = time.AfterFunc(t.after, t.execute)
	} else {
		t.ready = true
	}
}
