package worker

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type Task interface {
	Enqueue() error
	Start() error
	Stop() error
	Execute() error
	Error(err error) error
	Wait()
}

type Pool struct {
	sync.Mutex

	size  int
	tasks chan Task
	kill  chan struct{}
	wg    sync.WaitGroup
}

func NewPool(size int) *Pool {
	pool := &Pool{
		tasks: make(chan Task, 128),
		kill:  make(chan struct{}),
	}
	pool.Resize(size)
	return pool
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			log.Debugf("executingn task: %+v", task)
			err := task.Execute()
			if err != nil {
				log.Errorf("error executing task: %s", err)
				task.Error(err)
			}
		case <-p.kill:
			return
		}
	}
}

func (p *Pool) Resize(n int) {
	p.Lock()
	defer p.Unlock()
	for p.size < n {
		p.size++
		p.wg.Add(1)
		go p.worker()
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
