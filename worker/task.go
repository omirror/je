package worker

import (
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
