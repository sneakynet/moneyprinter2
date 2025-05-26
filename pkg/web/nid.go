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
		"types": map[string]types.NIDType{
			"SRI":      types.NIDTypeSRI,
			"Ethernet": types.NIDTypeEthernet,
		},
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

///////////////
// NID Ports //
///////////////

func (s *Server) uiViewNIDPortProvisionForm(w http.ResponseWriter, r *http.Request) {
	nidList, err := s.d.NIDList(&types.NID{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	account, err := s.d.AccountGet(&types.Account{ID: nidList[0].AccountID})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	equipment, err := s.d.EquipmentList(&types.Equipment{WirecenterID: nidList[0].Premise.WirecenterID})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ctx := pongo2.Context{
		"nid":       nidList[0],
		"account":   account,
		"equipment": equipment,
		"next":      r.URL.Query().Get("next"),
	}
	s.doTemplate(w, r, "views/nid/form_port.p2", ctx)
}

func (s *Server) uiViewNIDPortProvision(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	nidPort := types.NIDPort{
		ID:              s.strToUint(r.FormValue("nid_port_id")),
		NIDID:           s.strToUint(chi.URLParam(r, "id")),
		EquipmentPortID: s.strToUint(r.FormValue("equipment_port_id")),
	}

	if _, err := s.d.NIDPortSave(&nidPort); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	svcs := []types.Service{}
	for _, svcID := range r.Form["nid_services"] {
		svcs = append(svcs, types.Service{ID: s.strToUint(svcID)})
	}

	if err := s.d.NIDPortAssociateService(&nidPort, svcs); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, r.URL.Query().Get("next"), http.StatusSeeOther)
}
