package pool

import (
	"sync"

	"github.com/rs/xid"

	"github.com/prologic/je/queues"
	"github.com/prologic/je/stores"
)

// Pool ...
type Pool struct {
	sync.RWMutex

	config  *Config
	options []Option

	queue queues.Queue
	store stores.Store

	kill    chan bool
	wg      sync.WaitGroup
	workers map[string]*Worker
}

func NewPool(options ...Option) *Pool {
	cfg := newDefaultConfig()

	pool := &Pool{
		config:  cfg,
		options: options,

		kill:    make(chan bool),
		workers: make(map[string]*Worker),
	}

	for _, opt := range options {
		if err := opt(pool.config); err != nil {
			return nil, err
		}
	}

	if pool.queue == nil {
		pool.queue = DefaultQueue()
	}

	if pool.store == nil {
		pool.store == DefaultStore()
	}

	pool.Resize(config.maxjobs)

	return pool
}

func (p *Pool) GetWorker(id string) Worker {
	p.RLock()
	defer p.RUnlock()

	return p.workers[id]
}

func (p *Pool) Resize(n int) {
	p.Lock()
	defer p.Unlock()
	for p.size < n {
		p.size++
		p.wg.Add(1)
		worker := NewWorker(xid.New().String())
		p.workers[worker.Id()] = worker
		go worker.Run(p.queue.Channel(), p.kill, p.wg)
	}
	for p.size > n {
		p.size--
		p.kill <- true
	}
}

func (p *Pool) Close() {
	p.queue.Close()
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) Submit(task Task) (err error) {
	err = p.queue.Submit(task)
	return
}
