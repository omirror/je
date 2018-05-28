package je

import (
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
)

var db Store

func InitDB(uri string) (Store, error) {
	u, err := url.Parse(uri)
	if err != nil {
		log.Errorf("error parsing db uri %s: %s", uri, err)
		return nil, err
	}

	switch u.Scheme {
	case "memory":
		db, err = NewMemoryStore()
		if err != nil {
			log.Errorf("error creating store %s: %s", uri, err)
			return nil, err
		}
		return db, nil
	default:
		err := fmt.Errorf("unsupported db uri: %s", uri)
		log.Error(err)
		return nil, err
	}
}
