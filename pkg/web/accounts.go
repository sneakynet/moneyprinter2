package web

import (
	"net/http"
	"fmt"
	"log/slog"

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

	ctx := pongo2.Context{
		"account": account,
	}
	s.doTemplate(w, r, "views/account/detail.p2", ctx)
}

func (s *Server) uiViewAccountFormSingle(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "views/account/form_single.p2", nil)
}

func (s *Server) uiViewAccountFormBulk(w http.ResponseWriter, r *http.Request) {
	s.doTemplate(w, r, "views/account/form_bulk.p2", nil)
}

func (s *Server) uiAccountFormSubmitSingle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	aName := r.FormValue("account_name")
	aContact := r.FormValue("account_contact")
	aAlias := r.FormValue("account_alias")

	slog.Debug("Want to create account", "name", aName, "contact", aContact)

	id, err := s.d.AccountCreate(&types.Account{
		Name:    aName,
		Contact: aContact,
		Alias:   aAlias,
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
			_, err = s.d.AccountCreate(&types.Account{
				Name:    record["Name"],
				Contact: record["Contact"],
				Alias:   record["Alias"],
			})
			if err != nil {
				s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
				return
			}
		}
	}
	http.Redirect(w, r, "/ui/admin/accounts", http.StatusSeeOther)
}
