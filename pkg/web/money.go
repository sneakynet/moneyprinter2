package web

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/leekchan/accounting"

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
		types.FeeTargetCPE,
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
	lecs, err := s.d.LECList(&types.LEC{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	// This is messy, but this type assertion unviels the
	// interface to the database to the billing processor.
	// TODO(maldridge) clean this up.
	bp := billing.NewProcessor(billing.WithDatabase(s.d.(*db.DB)))
	if err := bp.Preload(lecs[0]); err != nil {
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
		bill, err := bp.BillAccount(account, lecs[0])
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
		width := s.strToUint(r.URL.Query().Get("width"))
		if width == 0 {
			width = 80
		}
		s.formatBillsText(w, bills, int(width))
	default:
		s.formatBillsHTML(w, bills)
	}
}

func (s *Server) uiViewBillForAccount(w http.ResponseWriter, r *http.Request) {
	accountID := s.strToUint(chi.URLParam(r, "id"))
	lecs, err := s.d.LECList(&types.LEC{ID: s.strToUint(chi.URLParam(r, "id"))})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	account, err := s.d.AccountGet(&types.Account{ID: accountID})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	// This is messy, but this type assertion unviels the
	// interface to the database to the billing processor.
	// TODO(maldridge) clean this up.
	bp := billing.NewProcessor(billing.WithDatabase(s.d.(*db.DB)))
	if err := bp.Preload(lecs[0]); err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	bill, err := bp.BillAccount(account, lecs[0])
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	switch r.Header.Get("Content-type") {
	case "text/plain":
		width := s.strToUint(r.URL.Query().Get("width"))
		if width == 0 {
			width = 80
		}
		s.formatBillsText(w, []billing.Bill{bill}, int(width))
	default:
		s.formatBillsHTML(w, []billing.Bill{bill})
	}
}

func (s *Server) formatBillsText(w http.ResponseWriter, bills []billing.Bill, width int) {
	for _, bill := range bills {
		t := table.NewWriter()
		t.SetColumnConfigs([]table.ColumnConfig{{
			Name:        "Service Bill",
			Align:       text.AlignCenter,
			AlignHeader: text.AlignCenter,
		}})
		t.Style().Options.DrawBorder = false
		t.Style().Size.WidthMin = width - 3
		t.SetOutputMirror(w)
		t.SetAllowedRowLength(width)
		t.AppendHeader(table.Row{"Service Bill"})
		t.AppendRow(table.Row{bill.LEC.Name + " - " + bill.LEC.Byline})
		t.AppendRow(table.Row{bill.LEC.Website})
		t.Render()
		fmt.Fprintln(w, "")

		t = table.NewWriter()
		t.Style().Options.DrawBorder = false
		t.Style().Size.WidthMin = width - 3
		t.SetOutputMirror(w)
		t.SetAllowedRowLength(width)
		t.AppendHeader(table.Row{"Name", "DBA", "Contact"})
		t.AppendRow(table.Row{bill.Account.Name, bill.Account.Alias, bill.Account.Contact})
		t.Render()
		fmt.Fprintln(w, "")

		t = table.NewWriter()
		t.SetColumnConfigs([]table.ColumnConfig{
			{
				Name:        "Cost",
				Align:       text.AlignRight,
				AlignHeader: text.AlignRight,
			},
		})
		t.Style().Options.DrawBorder = false
		t.Style().Size.WidthMin = width - 3
		t.SetOutputMirror(w)
		t.SetAllowedRowLength(width)
		t.AppendHeader(table.Row{"Fee", "Item", "Cost"})
		ac := accounting.Accounting{Symbol: "$", Precision: 2}
		for _, item := range bill.Lines {
			t.AppendRow(table.Row{item.Fee, item.Item, ac.FormatMoney(float64(item.Cost) / 100)})
		}
		t.SortBy([]table.SortBy{{Name: "Fee", Mode: table.Asc}})
		t.Render()
		fmt.Fprintln(w, "")

		t = table.NewWriter()
		t.Style().Options.DrawBorder = false
		t.SetOutputMirror(w)
		t.SetAllowedRowLength(width)
		t.AppendRow(table.Row{"Grand Total: " + ac.FormatMoney(float64(bill.Cost())/100)})
		t.Render()

		// Form feed, useful for line printers
		fmt.Fprintf(w, "%c", '\u000c')
	}
}

func (s *Server) formatBillsHTML(w http.ResponseWriter, bills []billing.Bill) {
	ctx := pongo2.Context{"bills": bills}
	s.doTemplate(w, nil, "views/bill/bills.p2", ctx)
}
