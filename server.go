package je

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	// Logging
	"github.com/unrolled/logger"

	// Routing
	"github.com/julienschmidt/httprouter"

	"github.com/prologic/je/pool"
)

const (
	DefaultDataPath = "./data"
	DefaultBacklog  = 32
	DefaultThreads  = 16
)

// Options ...
type Options struct {
	Data    string
	Backlog int
	Threads int
}

// Server ...
type Server struct {
	bind   string
	server *http.Server

	// Worker Pool
	pool *pool.Pool

	// Router
	router *httprouter.Router

	// Logger
	logger *logger.Logger
}

// ListenAndServe ...
func (s *Server) ListenAndServe() {
	log.Fatal(s.server.ListenAndServe())
}

func (s *Server) AddRoute(method, path string, handler http.Handler) {
	s.router.Handler(method, path, handler)
}

func (s *Server) Shutdown() {
	if err := s.server.Shutdown(context.Background()); err != nil {
		log.Errorf("error shutting down server: %v", err)
	}
}

func (s *Server) initRoutes() {
	s.router.GET("/", s.IndexHandler())
	s.router.POST("/create/*name", s.CreateHandler())
	s.router.POST("/kill/:id", s.KillHandler())
	s.router.GET("/logs/:id", s.LogsHandler())
	s.router.GET("/output/:id", s.OutputHandler())
	s.router.POST("/write/:id", s.WriteHandler())
	s.router.POST("/close/:id", s.CloseHandler())
	s.router.GET("/search", s.SearchHandler())
	s.router.GET("/search/:id", s.SearchHandler())
}

// NewServer ...
func NewServer(bind string, options *Options) *Server {
	var (
		backlog int
		threads int
	)

	if options != nil {
		backlog = options.Backlog
	} else {
		backlog = DefaultBacklog
	}

	if options != nil {
		threads = options.Threads
	} else {
		threads = DefaultThreads
	}

	router := httprouter.New()

	server := &Server{
		server: &http.Server{
			Addr: bind,
			Handler: logger.New(logger.Options{
				Prefix:               "je",
				RemoteAddressHeaders: []string{"X-Forwarded-For"},
			}).Handler(router),
		},

		// Worker Pool
		pool: pool.NewPool(
			pool.WithBacklog(backlog),
			pool.WithMaxJobs(thread),
		),

		// Router
		router: router,
	}

	server.initRoutes()

	return server
}
