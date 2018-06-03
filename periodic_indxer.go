package je

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/blevesearch/bleve"
)

type PeriodicIndexer struct {
	sync.Mutex
	bleveIndex

	interval time.Duration
	close    chan bool
	batch    *bleve.Batch
}

func NewPeriodicIndexer(index bleve.Index, interval time.Duration) bleve.Index {
	ib := &PeriodicIndexer{
		bleveIndex: index,

		batch:    index.NewBatch(),
		interval: interval,
		close:    make(chan bool),
	}

	go ib.batchloop()
	return ib
}

func (ib *PeriodicIndexer) batchloop() {
	t := time.NewTicker(ib.interval)

	for {
		select {
		case <-t.C:
			ib.Lock()
			err := ib.Batch(ib.batch)
			if err != nil {
				log.Errorf("error batching: %s", err)
			}
			ib.batch = ib.bleveIndex.NewBatch()
			ib.Unlock()
		case <-ib.close:
			break
		}
	}
}

func (ib *PeriodicIndexer) Close() error {
	ib.close <- true
	return ib.bleveIndex.Close()
}

func (ib *PeriodicIndexer) Index(id string, data interface{}) error {
	ib.Lock()
	err := ib.batch.Index(id, data)
	ib.Unlock()

	return err
}

func (ib *PeriodicIndexer) Delete(id string) error {
	ib.Lock()
	ib.batch.Delete(id)
	ib.Unlock()
	return nil
}

func (ib *PeriodicIndexer) SetInternal(key, val []byte) error {
	ib.Lock()
	ib.batch.SetInternal(key, val)
	ib.Unlock()
	return nil
}

func (ib *PeriodicIndexer) DeleteInternal(key []byte) error {
	ib.Lock()
	ib.batch.DeleteInternal(key)
	ib.Unlock()
	return nil
}
