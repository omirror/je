package je

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/blevesearch/bleve"
)

type MemoryStore struct {
	sync.RWMutex

	nextid ID
	data   map[ID]*Job
	index  bleve.Index
}

func (store *MemoryStore) Close() error {
	return nil
}

func (store *MemoryStore) NextId() ID {
	store.Lock()
	defer store.Unlock()

	store.nextid++
	return store.nextid
}

func (store *MemoryStore) Save(job *Job) error {
	store.Lock()
	store.data[job.ID] = job
	store.Unlock()

	store.index.Index(job.ID.String(), job)

	return nil
}

func (store *MemoryStore) Get(id ID) (job *Job, err error) {
	var ok bool

	store.RLock()
	job, ok = store.data[id]
	store.RUnlock()

	if !ok {
		err = &KeyError{id, ErrNotExist}
	}

	return
}

func (store *MemoryStore) Find(ids ...ID) (jobs []*Job, err error) {
	store.RLock()
	for _, id := range ids {
		job, ok := store.data[id]
		if ok {
			jobs = append(jobs, job)
		}
	}
	store.RUnlock()

	return
}

func (store *MemoryStore) All() (jobs []*Job, err error) {
	store.RLock()
	for _, job := range store.data {
		jobs = append(jobs, job)
	}
	store.RUnlock()

	return
}

func (store *MemoryStore) Search(q string) (jobs []*Job, err error) {
	query := bleve.NewQueryStringQuery(q)
	req := bleve.NewSearchRequest(query)
	res, err := store.index.Search(req)
	if err != nil {
		log.Errorf("error performing index search %s: %s", q, err)
		return
	}

	for _, hit := range res.Hits {
		store.RLock()
		job, ok := store.data[ParseId(hit.ID)]
		store.RUnlock()
		if !ok {
			log.Warnf("job #%s missing from store but exists in index!", hit.ID)
			continue
		}
		jobs = append(jobs, job)
	}

	return
}

func NewMemoryStore() (Store, error) {
	index, err := bleve.NewMemOnly(bleve.NewIndexMapping())
	if err != nil {
		log.Errorf("error creating index: %s", err)
		return nil, err
	}
	index = NewIndexBatcher(index, time.Millisecond*8)

	return &MemoryStore{
		data:  make(map[ID]*Job),
		index: index,
	}, nil
}