package web

import (
	"fmt"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

func (s *Server) uiViewNIDProvisionForm(w http.ResponseWriter, r *http.Request) {
	premises, err := s.d.PremiseList(&types.Premise{AccountID: s.strToUint(r.URL.Query().Get("account"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ctx := pongo2.Context{
		"accountID": 0,
		"premises":  premises,
		"types":     map[string]types.NIDType{"SRI": types.NIDTypeSRI},
	}

	s.doTemplate(w, r, "views/nid/form_single.p2", ctx)
}

func (s *Server) uiViewNIDProvision(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	premise, err := s.d.PremiseList(&types.Premise{ID: s.strToUint(r.FormValue("nid_premise"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	NID := types.NID{
		AccountID: s.strToUint(r.URL.Query().Get("account")),
		PremiseID: premise[0].ID,
		Type:      types.NIDType((s.strToUint(r.FormValue("nid_type")))),
		CLLI:      premise[0].Wirecenter.Name + premise[0].Address + r.URL.Query().Get("account"),
	}

	if _, err := s.d.NIDSave(&NID); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ui/admin/accounts/%d", NID.AccountID), http.StatusSeeOther)
}

func (s *Server) uiViewNIDDeprovision(w http.ResponseWriter, r *http.Request) {
	if err := s.d.NIDDelete(&types.NID{ID: s.strToUint(chi.URLParam(r, "id"))}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ui/admin/accounts/%s", r.URL.Query().Get("account")), http.StatusSeeOther)
}
