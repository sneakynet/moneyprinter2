package web

import (
	"os"
	"log/slog"
	"io/fs"
	"net/http"
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/flosch/pongo2/v6"
)

// Server handles the HTTP frontend.
type Server struct {
	r chi.Router
	n *http.Server

	tpl *pongo2.TemplateSet
}

// Option configures the Server
type Option func(*Server) error

// New returns a ready to serve webserver.
func New(opts ...Option) (*Server, error) {
	s := new(Server)
	s.r = chi.NewRouter()
	s.n = new(http.Server)

	var tplRoot fs.FS
	if tpath := os.Getenv("MONEYD_TEMPLATE_PATH"); tpath != "" {
		slog.Warn("Loading templates from debug path", "path", tpath)
		tplRoot = os.DirFS(tpath)
	} else {
		tplRoot, _ = fs.Sub(efs, "ui")
	}
	p2Root, _ := fs.Sub(tplRoot, "p2")
	ldr := pongo2.NewFSLoader(p2Root)
	s.tpl = pongo2.NewSet("html", ldr)
	_, s.tpl.Debug = os.LookupEnv("PONGO2_DEBUG")


	for _, o := range opts {
		o(s)
	}

	s.r.Use(middleware.Heartbeat("/ping"))

	s.r.Handle("/static/*", http.FileServer(http.FS(tplRoot)))

	s.r.Get("/", s.landing)
	s.r.Get("/login", s.login)

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

func (s *Server) landing(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "base.p2", nil)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "login.p2", nil)
}
