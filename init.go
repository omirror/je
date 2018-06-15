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
			return float64(time.Since(ctime).Nanoseconds())
		},
	)

	// server requests counter
	metrics.NewCounterVec(
		"server", "requests",
		"Number of requests made to the server",
		[]string{"method", "path"},
	)

	// job count counter
	metrics.NewCounter(
		"job", "count",
		"Job count",
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

	// job created gauge
	metrics.NewGauge(
		"job", "created",
		"Number of jobs in a created state",
	)
	// job waiting gauge
	metrics.NewGauge(
		"job", "waiting",
		"Number of jobs in a waiting state",
	)
	// job running gauge
	metrics.NewGauge(
		"job", "running",
		"Number of jobs in a running state",
	)
	// job killed gauge
	metrics.NewGauge(
		"job", "killed",
		"Number of jobs in a killed state",
	)
	// job stopped gauge
	metrics.NewGauge(
		"job", "stopped",
		"Number of jobs in a stopped state",
	)
	// job errored gauge
	metrics.NewGauge(
		"job", "errored",
		"Number of jobs in a errored state",
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
