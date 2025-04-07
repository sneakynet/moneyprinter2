package web

import (
	"fmt"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

func (s *Server) uiViewSwitchList(w http.ResponseWriter, r *http.Request) {
	switches, err := s.d.SwitchList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/switch/list.p2", pongo2.Context{"switches": switches})
}

func (s *Server) uiViewSwitchDetail(w http.ResponseWriter, r *http.Request) {
	switches, err := s.d.SwitchList(&types.Switch{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/switch/detail.p2", pongo2.Context{"switch": switches[0]})
}

func (s *Server) uiViewSwitchFormSingle(w http.ResponseWriter, r *http.Request) {
	lecs, err := s.d.LECList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	s.doTemplate(w, r, "views/switch/form_single.p2", pongo2.Context{"lecs": lecs})
}

func (s *Server) uiViewSwitchEdit(w http.ResponseWriter, r *http.Request) {
	lecs, err := s.d.LECList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	switches, err := s.d.SwitchList(&types.Switch{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ctx := pongo2.Context{
		"switch": switches[0],
		"lecs":   lecs,
	}

	s.doTemplate(w, r, "views/switch/form_single.p2", ctx)
}

func (s *Server) uiViewSwitchUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	sw := types.Switch{
		ID:    s.strToUint(chi.URLParam(r, "id")),
		CLLI:  r.FormValue("switch_clli"),
		Alias: r.FormValue("switch_alias"),
		LECID: s.strToUint(r.FormValue("switch_lec")),
	}

	id, err := s.d.SwitchSave(&sw)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ui/admin/switches/%d", id), http.StatusSeeOther)
}

func (s *Server) uiViewSwitchDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.d.SwitchDelete(&types.Switch{ID: s.strToUint(chi.URLParam(r, "id"))}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, "/ui/admin/switches", http.StatusSeeOther)
}
