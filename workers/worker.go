package workers

import (
	"io"
	"sync"

	"github.com/prologic/je"
)

type Worker interface {
	Id() string
	Kill(force bool) error
	CLose() error
	Write(input io.Reader) (int64, error)
	Run(tasks chan je.Task, kill chan bool, wg sync.WaitGroup)
}
