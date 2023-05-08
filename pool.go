package core

import (
	"sync"

	"github.com/gammazero/deque"
)

type QueuePool interface {
	Submit(key string, job func())
	Drain()
	Kill()
}

type QueueWorker interface {
	Submit(job func())
	Drain()
	Kill()
}

type queuePool struct {
	sync.Mutex

	workers   map[string]QueueWorker
	queueSize int
	drain     Fuse
	kill      Fuse
}

func NewQueuePool(queueSize int) QueuePool {
	return &queuePool{
		workers:   make(map[string]QueueWorker),
		queueSize: queueSize,
		drain:     NewFuse(),
		kill:      NewFuse(),
	}
}

func (p *queuePool) Submit(key string, job func()) {
	p.Lock()
	if p.kill.IsBroken() {
		p.Unlock()
		return
	}

	w, ok := p.workers[key]
	if !ok {
		w = NewQueueWorker(p.queueSize)
		p.workers[key] = w
	}
	p.Unlock()

	w.Submit(job)
}

func (p *queuePool) Drain() {
	p.drain.Once(func() {
		p.Lock()

		var wg sync.WaitGroup
		for _, w := range p.workers {
			wg.Add(1)
			w := w
			go func() {
				w.Drain()
				wg.Done()
			}()
		}
		wg.Wait()

		p.Unlock()
	})
}

func (p *queuePool) Kill() {
	p.kill.Once(func() {
		p.Lock()
		for _, w := range p.workers {
			w.Kill()
		}
		p.Unlock()
	})
}

type worker struct {
	sync.Mutex

	active   bool
	next     chan func()
	deque    *deque.Deque[func()]
	draining Fuse
	done     Fuse
	kill     Fuse
}

func NewQueueWorker(queueSize int) QueueWorker {
	w := &worker{
		next:     make(chan func(), 1),
		deque:    deque.New[func()](queueSize),
		draining: NewFuse(),
		done:     NewFuse(),
		kill:     NewFuse(),
	}
	go w.run()
	return w
}

func (w *worker) run() {
	kill := w.kill.Watch()
	for {
		select {
		case <-kill:
			return
		case job := <-w.next:
			job()

			w.Lock()
			if w.deque.Len() > 0 {
				w.next <- w.deque.PopFront()
			} else if w.draining.IsBroken() {
				w.done.Break()
				w.Unlock()
				return
			} else {
				w.active = false
			}
			w.Unlock()
		case <-w.draining.Watch():
			w.done.Break()
			return
		}
	}
}

func (w *worker) Submit(job func()) {
	w.Lock()
	if w.active {
		w.deque.PushBack(job)
	} else {
		w.active = true
		w.next <- job
	}
	w.Unlock()
}

func (w *worker) Drain() {
	w.draining.Break()
	select {
	case <-w.done.Watch():
	case <-w.kill.Watch():
	}
}

func (w *worker) Kill() {
	w.kill.Break()
}
