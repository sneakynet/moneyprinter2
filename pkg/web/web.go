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

	pongo2.RegisterFilter("key", s.filterGetValueByKey)

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

			r.Get("/{id}/manage-premises", s.uiViewAccountPremiseForm)
			r.Post("/{id}/manage-premises", s.uiViewAccountPremiseSubmit)
			r.Post("/{id}/unassign-premise", s.uiViewAccountPremiseUnassign)

			r.Get("/{id}/order-service", s.uiViewAccountServiceForm)
			r.Get("/{id}/order-service/{sid}", s.uiViewAccountServiceForm)
			r.Post("/{id}/order-service", s.uiViewAccountServiceUpsert)
			r.Post("/{id}/order-service/{sid}", s.uiViewAccountServiceUpsert)
			r.Post("/{id}/cancel-service/{sid}", s.uiViewAccountServiceCancel)
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

		a.Route("/services", func(r chi.Router) {
			r.Get("/", s.uiViewLECServiceList)

			r.Get("/new", s.uiViewLECServiceFormSingle)
			r.Post("/new", s.uiViewLECServiceUpsert)

			r.Get("/{id}/edit", s.uiViewLECServiceEdit)
			r.Post("/{id}/edit", s.uiViewLECServiceUpsert)
		})

		a.Route("/switches", func(r chi.Router) {
			r.Get("/", s.uiViewSwitchList)
			r.Get("/{id}", s.uiViewSwitchDetail)

			r.Get("/{id}/edit", s.uiViewSwitchEdit)
			r.Post("/{id}/edit", s.uiViewSwitchUpsert)

			r.Get("/{id}/config", s.uiViewSwitchConfig)

			r.Get("/new", s.uiViewSwitchFormSingle)
			r.Post("/new", s.uiViewSwitchUpsert)

			r.Post("/{id}/delete", s.uiViewSwitchDelete)

			r.Route("/{id}/equipment", func(r chi.Router) {
				r.Get("/", s.uiViewEquipmentList)
				r.Get("/{eid}", s.uiViewEquipmentDetail)

				r.Get("/{eid}/edit", s.uiViewEquipmentEdit)
				r.Post("/{eid}/edit", s.uiViewEquipmentUpsert)

				r.Get("/new", s.uiViewEquipmentFormSingle)
				r.Post("/new", s.uiViewEquipmentUpsert)

				r.Post("/{eid}/delete", s.uiViewEquipmentDelete)

				r.Route("/{eid}/ports", func(r chi.Router) {
					r.Get("/", s.uiViewPortList)

					r.Get("/{pid}/edit", s.uiViewPortEdit)
					r.Post("/{pid}/edit", s.uiViewPortUpsert)

					r.Get("/new", s.uiViewPortFormSingle)
					r.Post("/new", s.uiViewPortUpsert)

					r.Get("/bulk", s.uiViewPortFormBulk)
					r.Post("/bulk", s.uiViewPortFormBulkSubmit)

					r.Post("/{pid}/delete", s.uiViewPortDelete)
				})
			})
		})

		a.Route("/dn", func(r chi.Router) {
			r.Get("/", s.uiViewDNList)

			r.Get("/{id}/edit", s.uiViewDNEdit)
			r.Post("/{id}/edit", s.uiViewDNUpsert)

			r.Get("/new", s.uiViewDNFormSingle)
			r.Post("/new", s.uiViewDNUpsert)

			r.Get("/bulk", s.uiViewDNFormBulk)
			r.Post("/bulk", s.uiViewDNFormBulkSubmit)

			r.Post("/{id}/delete", s.uiViewDNDelete)
		})

		a.Route("/nid", func(r chi.Router) {
			r.Get("/provision", s.uiViewNIDProvisionForm)
			r.Post("/provision", s.uiViewNIDProvision)

			r.Post("/{id}/deprovision", s.uiViewNIDDeprovision)

			r.Route("/{id}/ports", func(r chi.Router) {
				r.Get("/provision", s.uiViewNIDPortProvisionForm)
				r.Post("/provision", s.uiViewNIDPortProvision)
			})
		})

		a.Route("/work", func(r chi.Router) {
			r.Get("/premises", s.uiViewWorkPremises)
			r.Get("/directory", s.uiViewWorkDirectory)
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
