package je

import (
	"fmt"
	"io"
)

// Task ...
type Task interface {
	Enqueue() error
	Start(worker string) error
	Stop() error
	Kill(force bool) error
	Killed() bool
	Close() error
	Write(input io.Reader) (int64, error)
	Execute() error
	Error(err error) error
	Wait()
}

// TaskQueue ...
type TaskQueue struct {
	q chan Task
}

func NewTaskQueue(buffer int) *TaskQueue {
	return &TaskQueue{
		q: make(chan Task, buffer),
	}
}

func (q *TaskQueue) Submit(task Task) error {
	select {
	case q.q <- task:
		//task.Enqueue()
		return nil
	default:
		return fmt.Errorf("queue is full or all workers are busy")
	}
}

func (q *TaskQueue) Channel() chan Task {
	return q.q
}

func (q *TaskQueue) Close() error {
	close(q.q)
	return nil
}
