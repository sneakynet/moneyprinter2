package web

import (
	"net/http"
	"sort"

	"github.com/flosch/pongo2/v6"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

func (s *Server) uiViewWorkPremises(w http.ResponseWriter, r *http.Request) {
	ctx := pongo2.Context{}

	type workOrder struct {
		Account  types.Account
		Premises types.Premise
		NIDs     []types.NID
	}
	workorders := []workOrder{}

	nids, err := s.d.NIDListFull(&types.NID{})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	nidMap := make(map[string][]types.NID)
	for _, nid := range nids {
		nidMap[nid.Premise.Address] = append(nidMap[nid.Premise.Address], nid)
	}
	for _, nids := range nidMap {
		workorders = append(workorders, workOrder{
			Account:  nids[0].Account,
			Premises: nids[0].Premise,
			NIDs:     nids,
		})
	}
	sort.Slice(workorders, func(i, j int) bool {
		return workorders[i].Account.Alias < workorders[j].Account.Alias
	})

	ctx["workorders"] = workorders

	s.doTemplate(w, r, "views/work/premise.p2", ctx)
}
