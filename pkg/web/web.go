package web

import (
	"context"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

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

	s.r.Route("/ui/admin", func(a chi.Router) {
		a.Route("/accounts", func(r chi.Router) {
			r.Get("/", s.uiViewAccountList)
			r.Get("/{id}", s.uiViewAccountDetail)
			r.Get("/new", s.uiViewAccountFormSingle)
			r.Get("/bulk", s.uiViewAccountFormBulk)

			r.Post("/new", s.uiAccountFormSubmitSingle)
			r.Post("/bulk", s.uiAccountFormSubmitBulk)
		})
		a.Route("/geo", func(b chi.Router) {
			b.Route("/wirecenters", func(r chi.Router) {
				r.Get("/", s.uiViewWirecenterList)
				r.Get("/{id}", s.uiViewWirecenterDetail)
				r.Get("/{id}/edit", s.uiViewWirecenterFormEdit)
				r.Post("/{id}/edit", s.uiViewWirecenterUpsert)

				r.Post("/{id}/delete", s.uiViewWirecenterDelete)

				r.Get("/new", s.uiViewWirecenterFormCreate)
				r.Post("/new", s.uiViewWirecenterUpsert)
			})

			b.Route("/premises", func(r chi.Router) {
				r.Get("/", s.uiViewPremisesList)
				r.Post("/{id}/delete", s.uiViewPremiseDelete)

				r.Get("/new", s.uiViewPremisesFormSingle)
				r.Post("/new", s.uiViewPremisesSubmitSingle)

				r.Get("/bulk", s.uiViewPremisesFormBulk)

				r.Post("/bulk", s.uiViewPremisesSubmitBulk)
			})
		})
		a.Route("/lecs", func(r chi.Router) {
			r.Get("/", s.uiViewLECList)
			r.Get("/{id}", s.uiViewLECDetail)

			r.Get("/{id}/edit", s.uiViewLECEdit)
			r.Post("/{id}/edit", s.uiViewLECUpsert)

			r.Get("/new", s.uiViewLECFormSingle)
			r.Post("/new", s.uiViewLECUpsert)
		})

		a.Route("/switches", func(r chi.Router) {
			r.Get("/", s.uiViewSwitchList)
			r.Get("/{id}", s.uiViewSwitchDetail)

			r.Get("/{id}/edit", s.uiViewSwitchEdit)
			r.Post("/{id}/edit", s.uiViewSwitchUpsert)

			r.Get("/new", s.uiViewSwitchFormSingle)
			r.Post("/new", s.uiViewSwitchUpsert)

			r.Post("/{id}/delete", s.uiViewSwitchDelete)
		})
	})

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
