package worker

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/rs/xid"
)

type Task interface {
	Id() uint64
	Enqueue() error
	Start(worker string) error
	Stop() error
	Kill(force bool) error
	Execute() error
	Error(err error) error
	Wait()
}

type Pool struct {
	sync.RWMutex

	size    int
	tasks   chan Task
	kill    chan struct{}
	wg      sync.WaitGroup
	workers map[string]*Worker
}

func NewPool(size int) *Pool {
	pool := &Pool{
		tasks:   make(chan Task, 128),
		kill:    make(chan struct{}),
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
		go worker.Run(p.tasks, p.kill, p.wg)
	}
	for p.size > n {
		p.size--
		p.kill <- struct{}{}
	}
}

func (p *Pool) Close() {
	close(p.tasks)
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) Submit(task Task) error {
	// TODO: Return an error if the task queue is full?
	p.tasks <- task
	task.Enqueue()
	log.Debugf("task enqueued for execution: %+v", task)
	return nil
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

func (w *Worker) Run(tasks chan Task, kill chan struct{}, wg sync.WaitGroup) {
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

			log.Debugf("executingn task: %+v", task)

			task.Start(w.Id())
			err := task.Execute()
			task.Stop()

			if err != nil {
				log.Errorf("error executing task: %s", err)
				task.Error(err)
			}
		case <-kill:
			return
		}
	}
}
