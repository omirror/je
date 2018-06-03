package je

import (
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
	bind string

	// Worker Pool
	pool *worker.Pool

	// Router
	router *httprouter.Router

	// Logger
	logger *logger.Logger
}

// ListenAndServe ...
func (s *Server) ListenAndServe() {
	log.Fatal(http.ListenAndServe(
		s.bind,
		s.logger.Handler(s.router),
	))
}

func (s *Server) initRoutes() {
	s.router.POST("/create/:name", s.CreateHandler())
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

	server := &Server{
		bind: bind,

		// Worker Pool
		pool: worker.NewPool(threads),

		// Router
		router: httprouter.New(),

		// Logger
		logger: logger.New(logger.Options{
			Prefix:               "je",
			RemoteAddressHeaders: []string{"X-Forwarded-For"},
		}),
	}

	server.initRoutes()

	return server
}
