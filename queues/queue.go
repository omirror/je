package queues

import (
	"github.com/prologic/je"
)

// Queue ...
type Queue interface {
	Submit(task je.Task) error
	Channel() chan je.Task
	Close() error
}
