package je

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	// Logging
	"github.com/unrolled/logger"

	// Stats/Metrics
	"github.com/rcrowley/go-metrics/exp"
	"github.com/thoas/stats"

	// Routing
	"github.com/julienschmidt/httprouter"

	"git.mills.io/prologic/je/worker"
)

const (
	DefaultWorkers = 16
)

// Options ...
type Options struct {
	Workers int
}

// Server ...
type Server struct {
	bind string

	// Worker Pool
	pool *worker.Pool

	// Router
	router *httprouter.Router

	// Logger
	logger *logger.Logger

	// Stats/Metrics
	counters *Counters
	stats    *stats.Stats
}

// ListenAndServe ...
func (s *Server) ListenAndServe() {
	log.Fatal(http.ListenAndServe(
		s.bind,
		s.logger.Handler(
			s.stats.Handler(s.router),
		),
	))
}

func (s *Server) initRoutes() {
	s.router.Handler("GET", "/debug/metrics", exp.ExpHandler(s.counters.r))
	s.router.GET("/debug/stats", s.StatsHandler())

	s.router.POST("/create/:name", s.CreateHandler())
	s.router.POST("/kill/:id", s.KillHandler())
	s.router.GET("/logs/:id", s.LogsHandler())
	s.router.GET("/output/:id", s.OutputHandler())
	s.router.GET("/search", s.SearchHandler())
	s.router.GET("/search/:id", s.SearchHandler())
}

// NewServer ...
func NewServer(bind string, options *Options) *Server {
	var (
		workers int
	)

	if options != nil {
		workers = options.Workers
	} else {
		workers = DefaultWorkers
	}

	server := &Server{
		bind: bind,

		// Worker Pool
		pool: worker.NewPool(workers),

		// Router
		router: httprouter.New(),

		// Logger
		logger: logger.New(logger.Options{
			Prefix:               "je",
			RemoteAddressHeaders: []string{"X-Forwarded-For"},
		}),

		// Stats/Metrics
		counters: NewCounters(),
		stats:    stats.New(),
	}

	server.initRoutes()

	return server
}
