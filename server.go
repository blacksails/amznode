package amznode

import (
	"net/http"

	"github.com/go-chi/chi"
)

// Server is the main type for instantiating the HTTP API
type Server interface {
	Storage() Storage
	Handler() http.Handler
}

type server struct {
	storage Storage
	r       *chi.Mux
}

// New instantiates a new amznode.Server
func New(storage Storage) Server {
	s := &server{
		storage: storage,
		r:       chi.NewMux(),
	}
	s.routes()
	return s
}

func (s *server) Storage() Storage {
	return s.storage
}

func (s *server) Handler() http.Handler {
	return s.r
}

func (s *server) routes() {
	r := s.r

	r.Post("/{childName}", s.createHandler())
	r.Post("/{parentID}/{childName}", s.createHandler())
	r.Get("/", s.getHandler())
	r.Get("/{id}", s.getHandler())
	r.Put("/{id}", s.changeParentHandler())
	r.Delete("/{id}", s.deleteHandler())
}
