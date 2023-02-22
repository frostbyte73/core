package core

import (
	"sync"
	"sync/atomic"
)

// Fuse is a thread-safe one-way switch, used for permanent state changes.
// Implementation partially borrowed from sync.Once
type Fuse interface {
	// IsOpen returns true if the fuse has not been broken
	IsOpen() bool
	// IsClosed returns true if the fuse has been broken
	IsClosed() bool
	// Wire returns a channel which will close once the fuse is broken
	Wire() <-chan struct{}
	// OnClose sets a callback to run when the fuse is broken
	OnClose(func())
	// Close breaks the fuse
	Close()
}

func NewFuse() Fuse {
	return &fuse{
		c: make(chan struct{}),
	}
}

type fuse struct {
	done    uint32
	m       sync.Mutex
	c       chan struct{}
	onClose func()
}

func (f *fuse) IsOpen() bool {
	select {
	case <-f.c:
		return false
	default:
		return true
	}
}

func (f *fuse) IsClosed() bool {
	select {
	case <-f.c:
		return true
	default:
		return false
	}
}

func (f *fuse) Wire() <-chan struct{} {
	return f.c
}

func (f *fuse) OnClose(onClose func()) {
	f.onClose = onClose
}

func (f *fuse) Close() {
	if atomic.LoadUint32(&f.done) == 0 {
		f.close()
	}
}

func (f *fuse) close() {
	f.m.Lock()
	defer f.m.Unlock()
	if f.done == 0 {
		defer atomic.StoreUint32(&f.done, 1)
		if onClose := f.onClose; onClose != nil {
			onClose()
		}
		close(f.c)
	}
}
