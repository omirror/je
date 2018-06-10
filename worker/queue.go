package worker

import (
	"fmt"
)

type Queue interface {
	Submit(task Task) error
	Channel() chan Task
	Close() error
}

type ChannelQueue struct {
	q chan Task
}

func NewChannelQueue(buffer int) *ChannelQueue {
	return &ChannelQueue{
		q: make(chan Task, buffer),
	}
}

func (q *ChannelQueue) Submit(task Task) error {
	select {
	case q.q <- task:
		task.Enqueue()
		return nil
	default:
		return fmt.Errorf("queue is full or all workers are busy")
	}
}

func (q *ChannelQueue) Channel() chan Task {
	return q.q
}

func (q *ChannelQueue) Close() error {
	close(q.q)
	return nil
}
