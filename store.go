package je

import (
	"fmt"
)

type ID uint64

func (id ID) String() string {
	return fmt.Sprintf("%d", id)
}

func ParseId(s string) ID {
	return ID(SafeParseUint64(s, 0))
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
