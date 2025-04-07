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
