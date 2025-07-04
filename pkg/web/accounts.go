package web

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

func (s *Server) uiViewAccountList(w http.ResponseWriter, r *http.Request) {
	accounts, err := s.d.AccountList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	s.doTemplate(w, r, "views/account/list.p2", pongo2.Context{"accounts": accounts})
}

func (s *Server) uiViewAccountDetail(w http.ResponseWriter, r *http.Request) {
	account, err := s.d.AccountGet(&types.Account{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	NIDs, err := s.d.NIDListFull(&types.NID{AccountID: account.ID})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ctx := pongo2.Context{
		"account": account,
		"nids":    NIDs,
	}
	s.doTemplate(w, r, "views/account/detail.p2", ctx)
}

func (s *Server) uiViewAccountFormSingle(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "views/account/form_single.p2", nil)
}

func (s *Server) uiViewAccountFormBulk(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "views/account/form_bulk.p2", nil)
}

func (s *Server) uiViewAccountEdit(w http.ResponseWriter, r *http.Request) {
	account, err := s.d.AccountGet(&types.Account{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	s.doTemplate(w, r, "views/account/form_single.p2", pongo2.Context{"account": account})
}

func (s *Server) uiAccountFormSubmitSingle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	id, err := s.d.AccountSave(&types.Account{
		ID:       s.strToUint(chi.URLParam(r, "id")),
		Name:     r.FormValue("account_name"),
		Contact:  r.FormValue("account_contact"),
		Alias:    r.FormValue("account_alias"),
		BillAddr: r.FormValue("account_billing"),
	})

	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ui/admin/accounts/%d", id), http.StatusSeeOther)
}

func (s *Server) uiAccountFormSubmitBulk(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	f, _, err := r.FormFile("accounts_file")
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	defer f.Close()
	records := s.csvToMap(f)

	for _, record := range records {
		if len(record["Name"]) == 0 {
			continue
		}

		_, err := s.d.AccountGet(&types.Account{Name: record["Name"]})
		if err != nil {
			slog.Warn("Error fetching account by name", "error", err)
			_, err = s.d.AccountSave(&types.Account{
				Name:     record["Name"],
				Contact:  record["Contact"],
				Alias:    record["Alias"],
				BillAddr: record["Billing"],
			})
			if err != nil {
				s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
				return
			}
		}
	}
	http.Redirect(w, r, "/ui/admin/accounts", http.StatusSeeOther)
}

func (s *Server) uiViewAccountPremiseForm(w http.ResponseWriter, r *http.Request) {
	account, err := s.d.AccountGet(&types.Account{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	wirecenters, err := s.d.WirecenterList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ctx := pongo2.Context{
		"account":     account,
		"wirecenters": wirecenters,
	}

	s.doTemplate(w, r, "views/account/premise.p2", ctx)
}

func (s *Server) uiViewAccountPremiseSubmit(w http.ResponseWriter, r *http.Request) {
	account, err := s.d.AccountGet(&types.Account{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	for _, p := range r.Form["account_premises"] {
		prem, err := s.d.PremiseList(&types.Premise{ID: s.strToUint(p)})
		if err != nil {
			s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
			return
		}
		prem[0].AccountID = account.ID
		slog.Debug("Binding premise to account", "account", account.ID, "premise", prem[0].ID)

		if _, err := s.d.PremiseSave(&prem[0]); err != nil {
			s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("/ui/admin/accounts/%d", account.ID), http.StatusSeeOther)
}

func (s *Server) uiViewAccountPremiseUnassign(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	premises, err := s.d.PremiseList(&types.Premise{ID: s.strToUint(r.FormValue("premise_id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	prem := premises[0]
	prem.AccountID = 0
	prem.Account = types.Account{}
	slog.Debug("Releasing premise", "premise", prem, "account", chi.URLParam(r, "id"))
	if _, err := s.d.PremiseSave(&prem); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/ui/admin/accounts/%s", chi.URLParam(r, "id")), http.StatusSeeOther)
}

func (s *Server) uiViewAccountServiceForm(w http.ResponseWriter, r *http.Request) {
	account, err := s.d.AccountGet(&types.Account{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	lecSvcs, err := s.d.LECServiceList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	lecServices := []struct {
		LEC      string
		Services []types.LECService
	}{}
	tmp := make(map[string][]types.LECService)
	for _, svc := range lecSvcs {
		tmp[svc.LEC.Name] = append(tmp[svc.LEC.Name], svc)
	}
	for lec, svc := range tmp {
		lecServices = append(lecServices, struct {
			LEC      string
			Services []types.LECService
		}{lec, svc})
	}

	availDN, err := s.d.DNListAvailable()
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	usedDN, err := s.d.DNListAssigned()
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	equipment, err := s.d.EquipmentList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	assignedPorts, err := s.d.PortListAssigned()
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	assigned := []uint{}
	for _, p := range assignedPorts {
		assigned = append(assigned, p.ID)
	}

	ctx := pongo2.Context{
		"Account":     account,
		"LECServices": lecServices,
		"AvailDN":     availDN,
		"UsedDN":      usedDN,
		"Equipment":   equipment,
		"Assigned":    assigned,
	}

	sid := s.strToUint(chi.URLParam(r, "sid"))
	if sid != 0 {
		slog.Debug("Querying for service orders", "sid", sid, "account", account.ID)
		svcs, err := s.d.ServiceList(&types.Service{
			ID:        sid,
			AccountID: account.ID,
		})
		if err != nil {
			slog.Debug("Could not retrieve service order", "error", err)
		}
		assignedDN := []uint{}
		if len(svcs) == 1 {
			for _, dn := range svcs[0].AssignedDN {
				assignedDN = append(assignedDN, dn.ID)
			}
			ctx["Order"] = svcs[0]
		}
		ctx["AssignedDN"] = assignedDN
		slog.Debug("Template Context", "ctx", assignedDN)
	}

	s.doTemplate(w, r, "views/account/order_service.p2", ctx)
}

func (s *Server) uiViewAccountServiceUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	svc := types.Service{
		ID:              s.strToUint(r.FormValue("service_id")),
		LECServiceID:    s.strToUint(r.FormValue("lec_service_id")),
		AccountID:       s.strToUint(chi.URLParam(r, "id")),
		EquipmentPortID: s.strToUint(r.FormValue("equipment_port_id")),
	}

	if _, err := s.d.ServiceSave(&svc); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	dns := []types.DN{}
	for _, dnID := range r.Form["assigned_dn"] {
		dns = append(dns, types.DN{ID: s.strToUint(dnID)})
	}
	if err := s.d.ServiceAssociateDN(&svc, dns); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ui/admin/accounts/%s", chi.URLParam(r, "id")), http.StatusSeeOther)
}

func (s *Server) uiViewAccountServiceCancel(w http.ResponseWriter, r *http.Request) {
	if err := s.d.ServiceDelete(&types.Service{ID: s.strToUint(chi.URLParam(r, "sid"))}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ui/admin/accounts/%s", chi.URLParam(r, "id")), http.StatusSeeOther)
}
