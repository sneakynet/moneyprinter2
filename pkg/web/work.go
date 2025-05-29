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

func (s *Server) uiViewWorkDirectory(w http.ResponseWriter, r *http.Request) {
	ctx := pongo2.Context{}

	type directoryEntry struct {
		Account types.Account
		DN      types.DN
	}
	directoryEntries := []directoryEntry{}

	svcs, err := s.d.ServiceList(&types.Service{})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	for _, svc := range svcs {
		for _, dn := range svc.AssignedDN {
			directoryEntries = append(directoryEntries, directoryEntry{
				Account: svc.Account,
				DN:      dn,
			})
		}
	}
	sort.Slice(directoryEntries, func(i, j int) bool {
		return directoryEntries[i].Account.Alias < directoryEntries[j].Account.Alias
	})

	ctx["directory"] = directoryEntries

	s.doTemplate(w, r, "views/work/directory.p2", ctx)
}
