package queues

import (
	"fmt"

	"github.com/prologic/je"
)

type ChannelQueue struct {
	q chan je.Task
}

func NewChannelQueue(buffer int) *ChannelQueue {
	return &ChannelQueue{
		q: make(chan je.Task, buffer),
	}
}

func (q *ChannelQueue) Submit(task je.Task) error {
	select {
	case q.q <- task:
		task.Enqueue()
		return nil
	default:
		return fmt.Errorf("queue is full or all workers are busy")
	}
}

func (q *ChannelQueue) Channel() chan je.Task {
	return q.q
}

func (q *ChannelQueue) Close() error {
	close(q.q)
	return nil
}
