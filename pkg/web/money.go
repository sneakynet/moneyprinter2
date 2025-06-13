package web

import (
	"log/slog"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"

	"github.com/sneakynet/moneyprinter2/pkg/billing"
	"github.com/sneakynet/moneyprinter2/pkg/db"
	"github.com/sneakynet/moneyprinter2/pkg/types"
)

func (s *Server) uiViewFeeList(w http.ResponseWriter, r *http.Request) {
	ctx := pongo2.Context{}
	fees, err := s.d.FeeList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	ctx["fees"] = fees
	s.doTemplate(w, r, "views/fee/list.p2", ctx)
}

func (s *Server) uiViewFeeFormSingle(w http.ResponseWriter, r *http.Request) {
	ctx := pongo2.Context{}

	lecs, err := s.d.LECList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	ctx["lecs"] = lecs
	ctx["targets"] = []types.FeeTarget{
		types.FeeTargetUnassigned,
		types.FeeTargetAccount,
		types.FeeTargetService,
		types.FeeTargetCPE,
		types.FeeTargetUsageCDR,
	}

	s.doTemplate(w, r, "views/fee/form_single.p2", ctx)
}

func (s *Server) uiViewFeeEdit(w http.ResponseWriter, r *http.Request) {
	ctx := pongo2.Context{}

	fees, err := s.d.FeeList(&types.Fee{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	ctx["fee"] = fees[0]

	lecs, err := s.d.LECList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	ctx["lecs"] = lecs
	ctx["targets"] = []types.FeeTarget{
		types.FeeTargetUnassigned,
		types.FeeTargetAccount,
		types.FeeTargetService,
		types.FeeTargetUsageCDR,
	}

	s.doTemplate(w, r, "views/fee/form_single.p2", ctx)
}

func (s *Server) uiViewFeeUpsert(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	fee := types.Fee{
		ID:         s.strToUint(chi.URLParam(r, "id")),
		Name:       r.FormValue("fee_name"),
		Expr:       r.FormValue("fee_expr"),
		Target:     types.FeeTarget(s.strToUint(r.FormValue("fee_target"))),
		LECReferer: s.strToUint(r.FormValue("assessed_by")),
	}

	if _, err := s.d.FeeSave(&fee); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, "/ui/admin/money/fees", http.StatusSeeOther)
}

func (s *Server) uiViewFeeDelete(w http.ResponseWriter, r *http.Request) {
	if err := s.d.FeeDelete(&types.Fee{ID: s.strToUint(chi.URLParam(r, "id"))}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	http.Redirect(w, r, "/ui/admin/money/fees", http.StatusSeeOther)
}

/////////////
// Billing //
/////////////

func (s *Server) uiViewBillList(w http.ResponseWriter, r *http.Request) {
	ctx := pongo2.Context{}

	accounts, err := s.d.AccountList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	for i := range accounts {
		// Fully hydrate the underlying account.
		accounts[i], _ = s.d.AccountGet(&accounts[i])
	}
	ctx["accounts"] = accounts

	lecs, err := s.d.LECList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	ctx["lecs"] = lecs

	s.doTemplate(w, r, "views/bill/list.p2", ctx)
}

func (s *Server) uiViewAllBillsForLEC(w http.ResponseWriter, r *http.Request) {
	lecID := s.strToUint(chi.URLParam(r, "id"))

	// This is messy, but this type assertion unviels the
	// interface to the database to the billing processor.
	// TODO(maldridge) clean this up.
	bp := billing.NewProcessor(billing.WithDatabase(s.d.(*db.DB)))
	if err := bp.Preload(types.LEC{ID: lecID}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	bills := []billing.Bill{}
	accounts, err := s.d.AccountList(nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	for i := range accounts {
		// Fully hydrate the underlying account.
		account, err := s.d.AccountGet(&types.Account{ID: accounts[i].ID})
		if err != nil {
			slog.Error("Potentially lost revenue while hydrating account", "account", account.ID, "error", err)
			continue
		}
		bill, err := bp.BillAccount(account, types.LEC{ID: lecID})
		if err != nil {
			slog.Error("Potentially lost revenue due to billing error", "account", account.ID, "error", err)
			continue
		}
		if bill.Cost() == 0 {
			continue
		}
		bills = append(bills, bill)
	}
	switch r.Header.Get("Content-type") {
	case "text/plain":
		s.formatBillsText(w, bills)
	default:
		s.formatBillsHTML(w, bills)
	}
}

func (s *Server) uiViewBillForAccount(w http.ResponseWriter, r *http.Request) {
	accountID := s.strToUint(chi.URLParam(r, "id"))
	lecID := s.strToUint(r.URL.Query().Get("lec"))

	account, err := s.d.AccountGet(&types.Account{ID: accountID})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	// This is messy, but this type assertion unviels the
	// interface to the database to the billing processor.
	// TODO(maldridge) clean this up.
	bp := billing.NewProcessor(billing.WithDatabase(s.d.(*db.DB)))
	if err := bp.Preload(types.LEC{ID: lecID}); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	bill, err := bp.BillAccount(account, types.LEC{ID: lecID})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	switch r.Header.Get("Content-type") {
	case "text/plain":
		s.formatBillsText(w, []billing.Bill{bill})
	default:
		s.formatBillsHTML(w, []billing.Bill{bill})
	}
}

func (s *Server) formatBillsText(w http.ResponseWriter, bills []billing.Bill) {

}

func (s *Server) formatBillsHTML(w http.ResponseWriter, bills []billing.Bill) {
	ctx := pongo2.Context{"bills": bills}
	s.doTemplate(w, nil, "views/bill/bills.p2", ctx)
}
