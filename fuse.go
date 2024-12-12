package core

import (
	"sync"
	"sync/atomic"
)

const (
	fuseNotInitialized = iota
	fuseReady
	fuseBroken
)

// Fuse is a thread-safe one-way switch, used for permanent state changes.
// Implementation partially borrowed from sync.Once
type Fuse struct {
	state uint32
	m     sync.Mutex
	c     chan struct{}
}

// IsBroken returns true if the fuse has been broken
func (f *Fuse) IsBroken() bool {
	switch atomic.LoadUint32(&f.state) {
	case fuseNotInitialized:
		return false
	case fuseBroken:
		return true
	default:
		select {
		case <-f.c:
			atomic.StoreUint32(&f.state, fuseBroken)
			return true
		default:
			return false
		}
	}
}

func (f *Fuse) init() {
	if atomic.LoadUint32(&f.state) != fuseNotInitialized {
		return
	}
	f.m.Lock()
	defer f.m.Unlock()
	if atomic.LoadUint32(&f.state) != fuseNotInitialized {
		return
	}
	f.c = make(chan struct{})
	atomic.StoreUint32(&f.state, fuseReady)
}

// Watch returns a channel which will close once the fuse is broken
func (f *Fuse) Watch() <-chan struct{} {
	f.init()
	return f.c
}

// Break breaks the fuse. It returns true if it was broken by this call.
func (f *Fuse) Break() bool {
	broken := false
	f.Once(func() {
		broken = true
	})
	return broken
}

// Once runs the callback and breaks the fuse
func (f *Fuse) Once(do func()) {
	if atomic.LoadUint32(&f.state) == fuseBroken {
		return
	}
	f.m.Lock()
	defer f.m.Unlock()
	switch atomic.LoadUint32(&f.state) {
	case fuseBroken:
		return
	case fuseNotInitialized:
		f.c = make(chan struct{})
		fallthrough
	default:
		defer atomic.StoreUint32(&f.state, fuseBroken)
		if do != nil {
			do()
		}
		close(f.c)
	}
}
