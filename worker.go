package je

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/prologic/msgbus"
	"github.com/prologic/msgbus/client"
	"github.com/prologic/observe"
	"github.com/rs/xid"
)

const (
	// DefaultBuffer ...
	DefaultBuffer = 16

	// DefaultWorkers ...
	DefaultWorkers = 32

	// DefaultJobType ...
	DefaultJobType = "*"
)

// BossOptions ...
type BossOptions struct {
	Buffer      int
	Workers     int
	JobType     string
	WithMetrics bool
}

// Boss ...
type Boss struct {
	sync.RWMutex

	metrics *observe.Metrics

	subscriber   *client.Subscriber
	loggerClient *client.Client
	queueClient  *client.Client

	jobType string

	size    int
	queue   *TaskQueue
	kill    chan bool
	wg      sync.WaitGroup
	workers map[string]*Worker
}

// NewBoss ...
func NewBoss(bind, logger, queue string, options *BossOptions) *Boss {
	var (
		buffer      int
		workers     int
		jobType     string
		withMetrics bool
	)

	if options != nil {
		buffer = options.Buffer
		workers = options.Workers
		jobType = options.JobType
		withMetrics = options.WithMetrics
	} else {
		buffer = DefaultBuffer
		workers = DefaultWorkers
		jobType = DefaultJobType
		withMetrics = false
	}

	var metrics *observe.Metrics

	if withMetrics {
		metrics = observe.NewMetrics("je worker")

		ctime := time.Now()

		// worker uptime counter
		metrics.NewCounterFunc(
			"worker", "uptime",
			"Number of nanoseconds the worker has been running",
			func() float64 {
				return float64(time.Since(ctime).Nanoseconds())
			},
		)

		// worker requests counter
		metrics.NewCounter(
			"worker", "requests",
			"Number of total requests processed",
		)

		/*
			// client latency summary
			metrics.NewSummary(
				"client", "latency_seconds",
				"Client latency in seconds",
			)

			// client errors counter
			metrics.NewCounter(
				"client", "errors",
				"Number of errors publishing messages to clients",
			)

			// bus messages counter
			metrics.NewCounter(
				"bus", "messages",
				"Number of total messages exchanged",
			)

			// bus dropped counter
			metrics.NewCounter(
				"bus", "dropped",
				"Number of messages dropped to subscribers",
			)

			// bus delivered counter
			metrics.NewCounter(
				"bus", "delivered",
				"Number of messages delivered to subscribers",
			)

			// bus fetched counter
			metrics.NewCounter(
				"bus", "fetched",
				"Number of messages fetched from clients",
			)

			// bus topics gauge
			metrics.NewCounter(
				"bus", "topics",
				"Number of active topics registered",
			)

			// queue len gauge vec
			metrics.NewGaugeVec(
				"queue", "len",
				"Queue length of each topic",
				[]string{"topic"},
			)

			// queue size gauge vec
			// TODO: Implement this gauge by somehow getting queue sizes per topic!
			metrics.NewGaugeVec(
				"queue", "size",
				"Queue length of each topic",
				[]string{"topic"},
			)

			// bus subscribers gauge
			metrics.NewGauge(
				"bus", "subscribers",
				"Number of active subscribers",
			)
		*/
	}

	boss := &Boss{
		metrics: metrics,

		subscriber:   nil,
		queueClient:  client.NewClient(queue, nil),
		loggerClient: client.NewClient(logger, nil),

		jobType: jobType,

		queue:   NewTaskQueue(buffer),
		kill:    make(chan bool),
		workers: make(map[string]*Worker),
	}
	boss.Resize(workers)
	boss.Start()
	return boss
}

// GetWorker ...
func (b *Boss) GetWorker(id string) *Worker {
	b.RLock()
	defer b.RUnlock()

	return b.workers[id]
}

// Start ...
func (b *Boss) Start() {
	b.subscriber = b.queueClient.Subscribe(b.jobType, b.handleMessage)
	b.subscriber.Start()
}

// Resize ...
func (b *Boss) Resize(n int) {
	b.Lock()
	defer b.Unlock()
	for b.size < n {
		b.size++
		b.wg.Add(1)
		worker := NewWorker(xid.New().String())
		b.workers[worker.ID()] = worker
		go worker.Run(b.queue.Channel(), b.kill, b.wg)
	}
	for b.size > n {
		b.size--
		b.kill <- true
	}
}

// Close ...
func (b *Boss) Close() {
	b.subscriber.Stop()
	b.queue.Close()
}

// Wait ...
func (b *Boss) Wait() {
	b.wg.Wait()
}

func (b *Boss) Shutdown() {
	b.Close()
	b.Wait()
}

func (b *Boss) Run() {
}

// Submit ...
func (b *Boss) Submit(task Task) (err error) {
	err = b.queue.Submit(task)
	return
}

// Metrics ...
func (b *Boss) Metrics() *observe.Metrics {
	return b.metrics
}

// ServeHTTP ...
func (b *Boss) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if b.metrics != nil {
			b.metrics.Counter("server", "requests").Inc()
		}
	}()

	if r.Method == "GET" && (r.URL.Path == "/" || r.URL.Path == "") {
		out, err := json.Marshal(b.workers)
		if err != nil {
			msg := fmt.Sprintf("error serializing workers: %s", err)
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	switch r.URL.Path {
	case "/write":
	case "/kill":
	case "/close":
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}

func (b *Boss) handleMessage(msg *msgbus.Message) error {
	var job Job

	log.Debugf("Payload: %s", msg.Payload)

	err := json.Unmarshal(msg.Payload, &job)
	if err != nil {
		log.Errorf("error decoding message payload: %s", err)
		return err
	}

	b.Submit(&job)

	return nil
}

// Worker ...
type Worker struct {
	sync.RWMutex

	id   string
	task Task
}

// NewWorker ...
func NewWorker(id string) *Worker {
	return &Worker{id: id}
}

// ID ...
func (w *Worker) ID() string {
	w.RLock()
	defer w.RUnlock()

	return w.id
}

// Kill ...
func (w *Worker) Kill(force bool) error {
	w.RLock()
	defer w.RUnlock()

	return w.task.Kill(force)
}

// Close ...
func (w Worker) Close() error {
	w.RLock()
	defer w.RUnlock()

	return w.task.Close()
}

// Write ...
func (w *Worker) Write(input io.Reader) (int64, error) {
	w.RLock()
	defer w.RUnlock()

	return w.task.Write(input)
}

// Run ...
func (w *Worker) Run(tasks chan Task, kill chan bool, wg sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				return
			}

			w.Lock()
			w.task = task
			w.Unlock()

			task.Start(w.ID())
			err := task.Execute()
			if err != nil {
				log.Errorf("error executing task: %s", err)
				task.Error(err)
			} else {
				if !task.Killed() {
					task.Stop()
				}
			}
		case <-kill:
			return
		}
	}
}
