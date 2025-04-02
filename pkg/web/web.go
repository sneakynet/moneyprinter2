package web

import (
	"net/http"
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server handles the HTTP frontend.
type Server struct {
	r chi.Router
	n *http.Server
}

// Option configures the Server
type Option func(*Server) error

// New returns a ready to serve webserver.
func New(opts ...Option) (*Server, error) {
	s := new(Server)
	s.r = chi.NewRouter()
	s.n = new(http.Server)

	for _, o := range opts {
		o(s)
	}

	s.r.Use(middleware.Heartbeat("/ping"))
	s.r.Get("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("hello")) })

	return s, nil
}

// Serve binds and serves http on the bound socket.  An error will be
// returned if the server cannot initialize.
func (s *Server) Serve(bind string) error {
	s.n.Addr = bind
	s.n.Handler = s.r
	return s.n.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.n.Shutdown(ctx)
}
