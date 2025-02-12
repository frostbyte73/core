package core

import (
	"hash/fnv"
	"sync"

	"github.com/gammazero/deque"
)

const (
	DefaultQueueSize = 100
	DefaultCapacity  = 10
)

type QueuePool interface {
	Submit(key string, job func()) bool
	Drain()
	Kill()
}

type QueueWorker interface {
	Submit(job func()) bool
	Drain()
	Kill()
}

type QueueWorkerParams struct {
	QueueSize    int
	DropWhenFull bool
}

type queuePool struct {
	sync.Mutex

	capacity int
	workers  []QueueWorker
	params   QueueWorkerParams
	drain    Fuse
	kill     Fuse
}

func NewQueuePool(maxWorkers int, params QueueWorkerParams) QueuePool {
	if maxWorkers <= 0 {
		maxWorkers = DefaultCapacity
	}
	if params.QueueSize == 0 {
		params.QueueSize = DefaultQueueSize
	}

	return &queuePool{
		capacity: maxWorkers,
		workers:  make([]QueueWorker, maxWorkers),
		params:   params,
	}
}

func (p *queuePool) hash(key string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return int(h.Sum32()) % p.capacity
}

func (p *queuePool) Submit(key string, job func()) bool {
	p.Lock()
	if p.kill.IsBroken() {
		p.Unlock()
		return false
	}

	idx := p.hash(key)
	w := p.workers[idx]
	if w == nil {
		w = NewQueueWorker(p.params)
		p.workers[idx] = w
	}
	p.Unlock()

	return w.Submit(job)
}

func (p *queuePool) Drain() {
	p.drain.Once(func() {
		p.Lock()

		var wg sync.WaitGroup
		for _, w := range p.workers {
			if w != nil {
				wg.Add(1)
				w := w
				go func() {
					w.Drain()
					wg.Done()
				}()
			}
		}
		wg.Wait()

		p.Unlock()
	})
}

func (p *queuePool) Kill() {
	p.kill.Once(func() {
		p.Lock()
		for _, w := range p.workers {
			if w != nil {
				w.Kill()
			}
		}
		p.Unlock()
	})
}

type worker struct {
	sync.Mutex
	QueueWorkerParams

	active   bool
	next     chan func()
	deque    deque.Deque[func()]
	draining Fuse
	done     Fuse
	kill     Fuse
}

func NewQueueWorker(params QueueWorkerParams) QueueWorker {
	if params.QueueSize == 0 {
		params.QueueSize = DefaultQueueSize
	}

	w := &worker{
		QueueWorkerParams: params,
		next:              make(chan func(), 1),
	}
	w.deque.SetBaseCap(params.QueueSize)
	go w.run()
	return w
}

func (w *worker) run() {
	kill := w.kill.Watch()
	draining := w.draining.Watch()
	for {
		select {
		case <-kill:
			return
		case <-draining:
			w.drain()
			return
		case job := <-w.next:
			if job != nil {
				job()
			}

			w.Lock()
			if w.deque.Len() > 0 {
				w.next <- w.deque.PopFront()
			} else {
				w.active = false
			}
			w.Unlock()
		}
	}
}

func (w *worker) Submit(job func()) bool {
	submitted := true
	w.Lock()
	if w.active {
		if w.DropWhenFull && w.deque.Len() == w.QueueSize {
			submitted = false
		} else {
			w.deque.PushBack(job)
		}
	} else {
		w.active = true
		w.next <- job
	}
	w.Unlock()
	return submitted
}

func (w *worker) Drain() {
	w.draining.Break()
	select {
	case <-w.done.Watch():
	case <-w.kill.Watch():
	}
}

func (w *worker) drain() {
	w.Lock()

	select {
	case job := <-w.next:
		if job != nil {
			job()
		}
	default:
	}
	count := w.deque.Len()
	for i := 0; i < count; i++ {
		w.deque.PopFront()()
	}

	w.done.Break()
	w.Unlock()
}

func (w *worker) Kill() {
	w.kill.Break()
}
