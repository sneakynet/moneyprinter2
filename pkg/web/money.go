package web

import (
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/go-chi/chi/v5"

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
