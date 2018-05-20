package je

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	// Database
	"github.com/asdine/storm"

	// Logging
	"github.com/unrolled/logger"

	// Stats/Metrics
	"github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"github.com/thoas/stats"

	// Routing
	"github.com/julienschmidt/httprouter"

	"git.mills.io/prologic/je/worker"
)

var (
	db *storm.DB
)

// Counters ...
type Counters struct {
	r metrics.Registry
}

func NewCounters() *Counters {
	counters := &Counters{
		r: metrics.NewRegistry(),
	}
	return counters
}

func (c *Counters) Inc(name string) {
	metrics.GetOrRegisterCounter(name, c.r).Inc(1)
}

func (c *Counters) Dec(name string) {
	metrics.GetOrRegisterCounter(name, c.r).Dec(1)
}

func (c *Counters) IncBy(name string, n int64) {
	metrics.GetOrRegisterCounter(name, c.r).Inc(n)
}

func (c *Counters) DecBy(name string, n int64) {
	metrics.GetOrRegisterCounter(name, c.r).Dec(n)
}

// Server ...
type Server struct {
	bind   string
	config Config
	router *httprouter.Router

	// Logger
	logger *logger.Logger

	// Stats/Metrics
	counters *Counters
	stats    *stats.Stats
}

// SearchJobsHandler ...
func (s *Server) SearchJobsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		s.counters.Inc("n_searchjobs")

		var jobs []*Job

		err := db.All(&jobs)
		if err != nil {
			log.Printf("error querying jobs index: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		out, err := json.Marshal(jobs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/json")
		w.Write(out)
	}
}

// NewJobHandler ...
func (s *Server) NewJobHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.counters.Inc("n_newjob")

		name := p.ByName("name")
		if name == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		job, err := NewJob(name)
		if err != nil {
			log.Printf("error creating new job: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		// TODO: Push new job to queue for workers
		err = job.Start()
		if err != nil {
			log.Printf("error starting job: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		res, err := worker.Run(name)
		if err != nil {
			log.Printf("error executing job: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		f, err := os.OpenFile(fmt.Sprintf("%d.log", job.ID), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Printf("error updates logs for job: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}
		f.Write(res.Logs())
		if err := f.Close(); err != nil {
			log.Printf("error closing logfile for job: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		// TODO: Persist job status, response and logs
		err = job.Finish(res)
		if err != nil {
			log.Printf("error updating job: %s", err)
			http.Error(w, "Internal Error", http.StatusInternalServerError)
			return
		}

		u, err := url.Parse(fmt.Sprintf("./%d", job.ID))
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		http.Redirect(w, r, r.URL.ResolveReference(u).String(), http.StatusFound)
	}
}

// ViewJobHandler ...
func (s *Server) ViewJobHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var job Job

		s.counters.Inc("n_viewjob")

		id := SafeParseInt(p.ByName("id"), 0)

		err := db.One("ID", id, &job)
		if err != nil && err == storm.ErrNotFound {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		out, err := json.Marshal(job)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/json")
		w.Write(out)
	}
}

// StatsHandler ...
func (s *Server) StatsHandler() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		bs, err := json.Marshal(s.stats.Data())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Write(bs)
	}
}

// ListenAndServe ...
func (s *Server) ListenAndServe() {
	log.Fatal(
		http.ListenAndServe(
			s.bind,
			s.logger.Handler(
				s.stats.Handler(s.router),
			),
		),
	)
}

func (s *Server) initRoutes() {
	s.router.Handler("GET", "/debug/metrics", exp.ExpHandler(s.counters.r))
	s.router.GET("/debug/stats", s.StatsHandler())

	s.router.GET("/job/:id", s.ViewJobHandler())
	s.router.POST("/job/:name", s.NewJobHandler())
	s.router.GET("/jobs", s.SearchJobsHandler())
}

// NewServer ...
func NewServer(bind string, config Config) *Server {
	server := &Server{
		bind:   bind,
		config: config,
		router: httprouter.New(),

		// Logger
		logger: logger.New(logger.Options{
			Prefix:               "je",
			RemoteAddressHeaders: []string{"X-Forwarded-For"},
			OutputFlags:          log.LstdFlags,
		}),

		// Stats/Metrics
		counters: NewCounters(),
		stats:    stats.New(),
	}

	server.initRoutes()

	return server
}
