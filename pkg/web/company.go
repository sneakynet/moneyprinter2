package web

import (
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

func (s *Server) uiViewLECList(w http.ResponseWriter, r *http.Request) {
	lecs, err := s.d.LECList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/lec/list.p2", pongo2.Context{"lecs": lecs})
}

func (s *Server) uiViewLECDetail(w http.ResponseWriter, r *http.Request) {
	lecs, err := s.d.LECList(&types.LEC{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/lec/detail.p2", pongo2.Context{"lec": lecs[0]})
}

func (s *Server) uiViewLECFormSingle(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "views/lec/form_single.p2", nil)
}

func (s *Server) uiViewLECEdit(w http.ResponseWriter, r *http.Request) {
	lecs, err := s.d.LECList(&types.LEC{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	s.doTemplate(w, r, "views/lec/form_single.p2", pongo2.Context{"lec": lecs[0]})
}

func (s *Server) uiViewLECUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	sw := types.LEC{
		ID:      s.strToUint(chi.URLParam(r, "id")),
		Name:    r.FormValue("lec_name"),
		Byline:  r.FormValue("lec_byline"),
		Contact: r.FormValue("lec_contact"),
		Website: r.FormValue("lec_website"),
	}

	_, err := s.d.LECSave(&sw)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, "/ui/admin/lecs", http.StatusSeeOther)
}

func (s *Server) uiViewLECServiceList(w http.ResponseWriter, r *http.Request) {
	svcs, err := s.d.LECServiceList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/svcs/list.p2", pongo2.Context{"svcs": svcs})
}

func (s *Server) uiViewLECServiceFormSingle(w http.ResponseWriter, r *http.Request) {
	lecs, err := s.d.LECList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	s.doTemplate(w, r, "views/svcs/form_single.p2", pongo2.Context{"lecs": lecs})
}

func (s *Server) uiViewLECServiceEdit(w http.ResponseWriter, r *http.Request) {
	lecs, err := s.d.LECList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	svcs, err := s.d.LECServiceList(&types.LECService{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	s.doTemplate(w, r, "views/svcs/form_single.p2", pongo2.Context{"lecs": lecs, "svc": svcs[0]})
}

func (s *Server) uiViewLECServiceUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	svc := types.LECService{
		ID:          s.strToUint(chi.URLParam(r, "id")),
		Name:        r.FormValue("service_name"),
		Slug:        r.FormValue("service_slug"),
		Description: r.FormValue("service_description"),
		LECID:       s.strToUint(r.FormValue("service_lec")),
	}

	if _, err := s.d.LECServiceSave(&svc); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, "/ui/admin/services/", http.StatusSeeOther)
}
