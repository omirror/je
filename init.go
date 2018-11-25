package je

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	queue   Queue
	store   Store
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

	return metrics
}

func InitQueue(uri string) (Queue, error) {
	u, err := url.Parse(uri)
	if err != nil {
		log.Errorf("error parsing queue uri %s: %s", uri, err)
		return nil, err
	}

	xs := strings.Split(u.Scheme, "+")
	if len(xs) == 2 {
		u.Scheme = xs[1]
	}

	switch xs[0] {
	case "msgbus":
		queue, err = NewMessageBusQueue(u.String())
		if err != nil {
			log.Errorf("error creating queue %s: %s", uri, err)
			return nil, err
		}
		log.Infof("Using MessageBusQueue %s", uri)
		return queue, nil
	default:
		err := fmt.Errorf("unsupported queue uri: %s", uri)
		log.Error(err)
		return nil, err
	}
}

func InitStore(uri string) (Store, error) {
	u, err := url.Parse(uri)
	if err != nil {
		log.Errorf("error parsing store uri %s: %s", uri, err)
		return nil, err
	}

	xs := strings.Split(u.Scheme, "+")
	if len(xs) == 2 {
		u.Scheme = xs[1]
	}

	switch xs[0] {
	case "memory":
		store, err = NewMemoryStore()
		if err != nil {
			log.Errorf("error creating store %s: %s", uri, err)
			return nil, err
		}
		log.Infof("Using MemoryStore %s", uri)
		return store, nil
	case "bolt":
		path := fmt.Sprintf("%s%s", u.Hostname(), u.EscapedPath())
		store, err = NewBoltStore(path)
		if err != nil {
			log.Errorf("error creating store %s: %s", uri, err)
			return nil, err
		}
		log.Infof("Using BoltStore %s", path)
		return store, nil
	case "je":
		store, err = NewRemoteStore(u.String())
		if err != nil {
			log.Errorf("error creating store %s: %s", uri, err)
			return nil, err
		}
		log.Infof("Using RemoteStore %s", u.String())
		return store, nil
	default:
		err := fmt.Errorf("unsupported store uri: %s", uri)
		log.Error(err)
		return nil, err
	}
}

// InitData returns a new Data object for persisting job logs and input
func InitData(uri string) (Data, error) {
	u, err := url.Parse(uri)
	if err != nil {
		log.Errorf("error parsing data uri %s: %s", uri, err)
		return nil, err
	}

	xs := strings.Split(u.Scheme, "+")
	if len(xs) == 2 {
		u.Scheme = xs[1]
	}

	switch xs[0] {
	case "file":
		path := fmt.Sprintf("%s%s", u.Hostname(), u.EscapedPath())
		data, err = NewLocalData(path)
		if err != nil {
			log.Errorf("error creating data %s: %s", path, err)
			return nil, err
		}
		log.Infof("Using LocalData %s", path)
		return data, nil
	case "je":
		data, err = NewRemoteData(u.String())
		if err != nil {
			log.Errorf("error creating store %s: %s", uri, err)
			return nil, err
		}
		log.Infof("Using RemoteData %s", u.String())
		return data, nil
	default:
		err := fmt.Errorf("unsupported data uri: %s", uri)
		log.Error(err)
		return nil, err
	}
}
