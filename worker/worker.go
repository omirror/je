package worker

import (
	"io"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/rs/xid"
)

// Pool ...
type Pool struct {
	sync.RWMutex

	size    int
	queue   Queue
	kill    chan bool
	wg      sync.WaitGroup
	workers map[string]*Worker
}

func NewPool(backlog, size int) *Pool {
	pool := &Pool{
		queue:   NewChannelQueue(backlog),
		kill:    make(chan bool),
		workers: make(map[string]*Worker),
	}
	pool.Resize(size)
	return pool
}

func (p *Pool) GetWorker(id string) *Worker {
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

type Worker struct {
	sync.RWMutex

	id   string
	task Task
}

func NewWorker(id string) *Worker {
	return &Worker{id: id}
}

func (w *Worker) Id() string {
	w.RLock()
	defer w.RUnlock()

	return w.id
}

func (w *Worker) Kill(force bool) error {
	w.RLock()
	defer w.RUnlock()

	return w.task.Kill(force)
}

func (w Worker) Close() error {
	w.RLock()
	defer w.RUnlock()

	return w.task.Close()
}

func (w *Worker) Write(input io.Reader) (int64, error) {
	w.RLock()
	defer w.RUnlock()

	return w.task.Write(input)
}

func (w *Worker) Run(tasks chan Task, kill chan bool, wg sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				return
			}

			w.Lock()
			w.task = task
			w.Unlock()

			task.Start(w.Id())
			err := task.Execute()
			if err != nil {
				log.Errorf("error executing task: %s", err)
				task.Error(err)
			} else {
				if !task.Killed() {
					task.Stop()
				}
			}
		case <-kill:
			return
		}
	}
}
