package web

import (
	"log/slog"
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

func (s *Server) uiViewPremisesList(w http.ResponseWriter, r *http.Request) {
	premises, err := s.d.PremiseList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/premise/list.p2", pongo2.Context{"premises": premises})
}

func (s *Server) uiViewPremisesFormSingle(w http.ResponseWriter, r *http.Request) {
	wirecenters, err := s.d.WirecenterList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	s.doTemplate(w, r, "views/premise/form_single.p2", pongo2.Context{"wirecenters": wirecenters})
}

func (s *Server) uiViewPremisesSubmitSingle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	pAddress := r.FormValue("premise_address")
	pAlias := r.FormValue("premise_alias")
	pWirecenterID := s.strToUint(r.FormValue("premise_wirecenter"))

	slog.Debug("Want to create premise", "address", pAddress, "alias", pAlias, "wirecenter", pWirecenterID)

	_, err := s.d.PremiseSave(&types.Premise{
		Address:      pAddress,
		WirecenterID: pWirecenterID,
	})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, "/ui/admin/geo/premises", http.StatusSeeOther)
}

func (s *Server) uiViewPremisesFormBulk(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "views/premise/form_bulk.p2", nil)
}

func (s *Server) uiViewPremisesSubmitBulk(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	f, _, err := r.FormFile("premises_file")
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	defer f.Close()
	records := s.csvToMap(f)

	wirecenters, err := s.d.WirecenterList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	wc := make(map[string]uint)
	for _, w := range wirecenters {
		wc[w.Name] = w.ID
	}

	for _, record := range records {
		if len(record["Address"]) == 0 {
			continue
		}

		res, err := s.d.PremiseList(&types.Premise{Address: record["Address"]})
		if len(res) == 0 {
			slog.Info("Premise did not exist, creating", "address", record["Address"], "alias", record["Alias"], "wirecenter", wc[record["Wirecenter"]])
			_, err = s.d.PremiseSave(&types.Premise{
				Address:      record["Address"],
				WirecenterID: wc[record["Wirecenter"]],
			})
			if err != nil {
				s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
				return
			}
		}
	}
	http.Redirect(w, r, "/ui/admin/geo/premises", http.StatusSeeOther)
}

func (s *Server) uiViewPremiseDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.d.PremiseDelete(&types.Premise{ID: s.strToUint(chi.URLParam(r, "id"))}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, "/ui/admin/geo/premises", http.StatusSeeOther)
}
