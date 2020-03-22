package workers

import (
	"io"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/prologic/je"
)

type LocalWorker struct {
	sync.RWMutex

	id   string
	task je.Task
}

func NewLocalWorker(id string) *LocalWorker {
	return &LocalWorker{id: id}
}

func (w *LocalWorker) Id() string {
	w.RLock()
	defer w.RUnlock()

	return w.id
}

func (w *LocalWorker) Kill(force bool) error {
	w.RLock()
	defer w.RUnlock()

	return w.task.Kill(force)
}

func (w *LocalWorker) Close() error {
	w.RLock()
	defer w.RUnlock()

	return w.task.Close()
}

func (w *LocalWorker) Write(input io.Reader) (int64, error) {
	w.RLock()
	defer w.RUnlock()

	return w.task.Write(input)
}

func (w *LocalWorker) Run(tasks chan Task, kill chan bool, wg sync.WaitGroup) {
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
