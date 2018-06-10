package je

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	db      Store
	data    Data
	metrics *Metrics
)

func InitMetrics(name string) *Metrics {
	metrics = NewMetrics(name)

	ctime := time.Now()

	// server uptime counter
	metrics.NewCounterFunc(
		"server", "uptime",
		"Number of nanoseconds the server has been running",
		func() float64 {
			return float64(time.Since(ctime).NanoSeconds())
		},
	)

	// server requests counter
	metrics.NewCounterVec(
		"server", "requests",
		"Number of requests made to the server",
		[]string{"method", "path"},
	)

	// job duration summary
	metrics.NewSummaryVec(
		"job", "duration",
		"Job duration in seconds",
		[]string{"name"},
	)

	// job index summary
	metrics.NewSummary(
		"job", "index",
		"Index duration in seconds",
	)

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

func InitData(path string) (Data, error) {
	var err error

	data, err = NewLocalData(path)

	return data, err
}
