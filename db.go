package je

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

var db Store

type URI struct {
	Type string
	Path string
}

func (u *URI) String() string {
	return fmt.Sprintf("%s://%s", u.Type, u.Path)
}

func ParseURI(uri string) (*URI, error) {
	parts := strings.Split(uri, "://")
	if len(parts) == 2 {
		return &URI{Type: strings.ToLower(parts[0]), Path: parts[1]}, nil
	}
	return nil, fmt.Errorf("invalid uri: %s", uri)
}

func InitDB(uri string) (Store, error) {
	u, err := ParseURI(uri)
	if err != nil {
		log.Errorf("error parsing db uri %s: %s", uri, err)
		return nil, err
	}

	switch u.Type {
	case "memory":
		db, err = NewMemoryStore()
		if err != nil {
			log.Errorf("error creating store %s: %s", uri, err)
			return nil, err
		}
		log.Infof("Using MemoryStore %s", uri)
		return db, nil
	case "bolt":
		db, err = NewBoltStore(u.Path)
		if err != nil {
			log.Errorf("error creating store %s: %s", uri, err)
			return nil, err
		}
		log.Infof("Using BoltStore %s", uri)
		return db, nil
	default:
		err := fmt.Errorf("unsupported db uri: %s", uri)
		log.Error(err)
		return nil, err
	}
}
