package je

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrNotExist = errors.New("key does not exist")
)

type KeyError struct {
	Key ID
	Err error
}

func (e *KeyError) Error() string {
	return fmt.Sprintf("%s: %d", e.Err, e.Key)
}

type ID uint64

func (id ID) String() string {
	return fmt.Sprintf("%d", id)
}

func (id ID) Bytes() []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(id))
	return b
}

func ParseId(s string) ID {
	return ID(SafeParseUint64(s, 0))
}

type IdGenerator struct {
	sync.Mutex
	next ID
}

func (id *IdGenerator) Next() ID {
	id.Lock()
	defer id.Unlock()

	id.next++
	return id.next
}

type Store interface {
	Close() error
	NextId() ID
	Save(job *Job) error
	Get(id ID) (*Job, error)
	Find(id ...ID) ([]*Job, error)
	All() ([]*Job, error)
	Search(q string) ([]*Job, error)
}
