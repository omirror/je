package je

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	// Logging
	"github.com/unrolled/logger"

	// Routing
	"github.com/julienschmidt/httprouter"

	"git.mills.io/prologic/je/worker"
)

const (
	DefaultDataPath = "./data"
	DefaultThreads  = 16
)

// Options ...
type Options struct {
	Data    string
	Threads int
}

// Server ...
type Server struct {
	bind   string
	server *http.Server

	// Worker Pool
	pool *worker.Pool

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
		threads int
	)

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
		pool: worker.NewPool(threads),

		// Router
		router: router,
	}

	server.initRoutes()

	return server
}
