package je

//go:generate rice embed-go

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/julienschmidt/httprouter"
	"github.com/unrolled/logger"
)

// Server ...
type Server struct {
	bind   string
	server *http.Server

	// Data
	data Data

	// Queue
	queue Queue

	// Store
	store Store

	// Router
	router *httprouter.Router

	// Logger
	logger *logger.Logger
}

// ListenAndServe ...
func (s *Server) ListenAndServe() {
	log.Fatal(s.server.ListenAndServe())
}

func (s *Server) GetWorker(id string) *Worker {
	return &Worker{}
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
	s.router.GET("/read/:id", s.ReadHandler())
	s.router.POST("/write/:id", s.WriteHandler())
	s.router.POST("/write/:id/:dtype", s.WriteHandler())
	s.router.POST("/close/:id", s.CloseHandler())
	s.router.GET("/search", s.SearchHandler())
	s.router.GET("/search/:id", s.SearchHandler())
	s.router.POST("/update/:id", s.UpdateHandler())
}

// NewServer ...
func NewServer(bind string, data Data, queue Queue, store Store) *Server {
	router := httprouter.New()

	server := &Server{
		server: &http.Server{
			Addr: bind,
			Handler: logger.New(logger.Options{
				Prefix:               "je",
				RemoteAddressHeaders: []string{"X-Forwarded-For"},
			}).Handler(router),
		},

		data:   data,
		queue:  queue,
		store:  store,
		router: router,
	}

	server.initRoutes()

	return server
}
