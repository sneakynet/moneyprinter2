package web

import (
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

func (s *Server) uiViewWirecenterList(w http.ResponseWriter, r *http.Request) {
	wirecenters, err := s.d.WirecenterList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/wirecenter/list.p2", pongo2.Context{"wirecenters": wirecenters})
}

func (s *Server) uiViewWirecenterDetail(w http.ResponseWriter, r *http.Request) {
	wirecenter, err := s.d.WirecenterGet(&types.Wirecenter{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/wirecenter/detail.p2", pongo2.Context{"wirecenter": wirecenter})
}

func (s *Server) uiViewWirecenterUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	sw := types.Wirecenter{
		ID:   s.strToUint(chi.URLParam(r, "id")),
		Name: r.FormValue("wirecenter_name"),
	}

	_, err := s.d.WirecenterSave(&sw)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, "/ui/admin/geo/wirecenters", http.StatusSeeOther)
}

func (s *Server) uiViewWirecenterFormCreate(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "views/wirecenter/form.p2", nil)
}

func (s *Server) uiViewWirecenterFormEdit(w http.ResponseWriter, r *http.Request) {
	wc, err := s.d.WirecenterGet(&types.Wirecenter{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	s.doTemplate(w, r, "views/wirecenter/form.p2", pongo2.Context{"wirecenter": wc})
}

func (s *Server) uiViewWirecenterDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.d.WirecenterDelete(&types.Wirecenter{ID: s.strToUint(chi.URLParam(r, "id"))}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, "/ui/admin/geo/wirecenters", http.StatusSeeOther)
}
