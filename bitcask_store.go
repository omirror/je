package je

import (
	"fmt"
	"os"
	"path"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/blevesearch/bleve"
	"github.com/prologic/bitcask"

	"github.com/prologic/je/codec"
	"github.com/prologic/je/codec/json"
)

type BitcaskStore struct {
	db     *bitcask.Bitcask
	nextid *IdGenerator
	index  bleve.Index
	codec  codec.MarshalUnmarshaler
}

func (store *BitcaskStore) Close() error {
	return store.db.Close()
}

func (store *BitcaskStore) NextId() ID {
	return ID(0)
}

func (store *BitcaskStore) Save(job *Job) error {
	if job.ID == ID(0) {
		job.ID = store.nextid.Next()
	}

	val, err := store.codec.Marshal(job)
	if err != nil {
		log.Errorf("error serializing job: %s", err)
		return err
	}

	key := []byte(fmt.Sprintf("job_%d", job.ID))

	if err := store.db.Put(key, val); err != nil {
		log.Errorf("error saving job: %s", err)
		return err
	}

	t := time.Now()
	store.index.Index(job.ID.String(), job)
	metrics.Summary("job", "index").Observe(time.Now().Sub(t).Seconds())

	return nil
}

func (store *BitcaskStore) Get(id ID) (job *Job, err error) {
	key := []byte(fmt.Sprintf("job_%s", id))
	val, err := store.db.Get(key)
	if err != nil {
		if err == bitcask.ErrKeyNotFound {
			log.Errorf("job #%d not found", id)
			err = &KeyError{id, ErrNotExist}
			return
		}
		log.Error("error feteching job #%d : %s", id, err)
		return
	}

	err = store.codec.Unmarshal(val, &job)
	if err != nil {
		log.Errorf("error deserializing job #%s: %s", id, err)
		return
	}

	return
}

func (store *BitcaskStore) Find(ids ...ID) (jobs []*Job, err error) {
	jobs = make([]*Job, len(ids))
	for i, id := range ids {
		key := []byte(fmt.Sprintf("job_%d", id))
		val, err := store.db.Get(key)
		if err != nil {
			return nil, err
		}

		if err := store.codec.Unmarshal(val, &jobs[i]); err != nil {
			log.Errorf("error deserializing job #%s: %s", id, err)
			return nil, err
		}
	}
	return
}

func (store *BitcaskStore) All() (jobs []*Job, err error) {
	prefix := []byte("job_")
	err = store.db.Scan(prefix, func(key []byte) error {
		var job Job

		val, err := store.db.Get(key)
		if err != nil {
			log.Error("error fetching job %s : %s", string(key), err)
			return err
		}
		if err := store.codec.Unmarshal(val, &job); err != nil {
			log.Errorf("error deserializing jobs: %s", err)
			return err
		}

		jobs = append(jobs, &job)
		return nil
	})
	return
}

func (store *BitcaskStore) Search(q string) (jobs []*Job, err error) {
	size, err := store.index.DocCount()
	if err != nil {
		log.Errorf("error getting index size: %s", err)
		return
	}

	query := bleve.NewQueryStringQuery(q)
	req := bleve.NewSearchRequestOptions(query, int(size), 0, false)
	res, err := store.index.Search(req)
	if err != nil {
		log.Errorf("error performing index search %s: %s", q, err)
		return
	}

	var ids []ID
	for _, hit := range res.Hits {
		ids = append(ids, ParseId(hit.ID))
	}

	jobs, err = store.Find(ids...)

	return
}

func NewBitcaskStore(dbpath string) (Store, error) {
	db, err := bitcask.Open(dbpath)
	if err != nil {
		log.Errorf("error opening store %s: %s", dbpath, err)
		return nil, err
	}

	var index bleve.Index
	indexpath := path.Join(path.Dir(dbpath), "index.db")
	if _, err = os.Stat(indexpath); err == nil {
		index, err = bleve.Open(indexpath)
	} else {
		index, err = bleve.New(indexpath, bleve.NewIndexMapping())
	}
	if err != nil {
		log.Errorf("error creating index: %s", err)
		return nil, err
	}

	return &BitcaskStore{
		db:     db,
		nextid: &IdGenerator{},
		index:  index,
		codec:  json.Codec,
	}, nil
}
