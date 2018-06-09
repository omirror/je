package je

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

var (
	db      Store
	metrics *Metrics
)

func InitMetrics(name string) *Metrics {
	metrics = NewMetrics(name)

	return metrics
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
