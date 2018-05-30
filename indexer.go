package je

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/blevesearch/bleve"
)

type bleveIndex bleve.Index

type IndexBatcher struct {
	sync.Mutex
	bleveIndex

	interval time.Duration
	close    chan bool
	batch    *bleve.Batch
}

func NewIndexBatcher(index bleve.Index, interval time.Duration) bleve.Index {
	ib := &IndexBatcher{
		bleveIndex: index,

		batch:    index.NewBatch(),
		interval: interval,
		close:    make(chan bool),
	}

	go ib.batchloop()
	return ib
}

func (ib *IndexBatcher) batchloop() {
	t := time.NewTicker(ib.interval)

	for {
		select {
		case <-t.C:
			ib.Lock()
			err := ib.Batch(ib.batch)
			if err != nil {
				log.Errorf("error batching: %s", err)
			}
			ib.batch.Reset()
			ib.Unlock()
		case <-ib.close:
			break
		}
	}
}

func (ib *IndexBatcher) Close() error {
	ib.close <- true
	return ib.bleveIndex.Close()
}

func (ib *IndexBatcher) Index(id string, data interface{}) error {
	ib.Lock()
	err := ib.batch.Index(id, data)
	ib.Unlock()

	return err
}

func (ib *IndexBatcher) Delete(id string) error {
	ib.Lock()
	ib.batch.Delete(id)
	ib.Unlock()
	return nil
}

func (ib *IndexBatcher) SetInternal(key, val []byte) error {
	ib.Lock()
	ib.batch.SetInternal(key, val)
	ib.Unlock()
	return nil
}

func (ib *IndexBatcher) DeleteInternal(key []byte) error {
	ib.Lock()
	ib.batch.DeleteInternal(key)
	ib.Unlock()
	return nil
}
