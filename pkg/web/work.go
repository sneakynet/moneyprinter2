package web

import (
	"log/slog"
	"net/http"
	"sort"

	"github.com/flosch/pongo2/v6"

	"github.com/sneakynet/moneyprinter2/pkg/billing"
	"github.com/sneakynet/moneyprinter2/pkg/db"
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

	nids, err := s.d.NIDListFull(r.Context(), &types.NID{})
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

	svcs, err := s.d.ServiceList(r.Context(), &types.Service{})
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

func (s *Server) uiViewWorkStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := pongo2.Context{}

	type serviceCount struct {
		Service string
		Slug    string
		Count   int
	}

	svcs, err := s.d.ServiceList(r.Context(), &types.Service{})
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	counts := make(map[uint]*serviceCount)
	for _, svc := range svcs {
		if existing, ok := counts[svc.LECServiceID]; ok {
			existing.Count++
		} else {
			counts[svc.LECServiceID] = &serviceCount{
				Service: svc.LECService.Name,
				Slug:    svc.LECService.Slug,
				Count:   1,
			}
		}
	}

	slice := make([]serviceCount, 0, len(counts))
	for _, c := range counts {
		slice = append(slice, *c)
	}
	sort.Slice(slice, func(i, j int) bool {
		return slice[i].Count > slice[j].Count
	})

	ctx["services"] = slice

	type cdrStat struct {
		Value string
		CNAM  string
		Count int
	}

	type cdrStatMinutes struct {
		Value         string
		CNAM          string
		TotalSeconds  float64
		LongestSecond float64
	}

	billableSub := s.d.Raw().Model(&types.CDR{}).
		Select("DISTINCT cdrs.id").
		Joins("INNER JOIN switches ON cdrs.clli = switches.clli").
		Joins("INNER JOIN dns ON cdrs.clid = dns.number").
		Joins("INNER JOIN dn_assignments ON dns.id = dn_assignments.dn_id").
		Joins("INNER JOIN services ON dn_assignments.service_id = services.id").
		Joins("INNER JOIN ports ON services.equipment_port_id = ports.id").
		Joins("INNER JOIN equipment ON ports.equipment_id = equipment.id").
		Where("switches.id = equipment.switch_id")

	var origins []cdrStat
	if err := s.d.Raw().Model(&types.CDR{}).
		Select("cdrs.clid AS value, dns.cnam AS cnam, COUNT(cdrs.id) AS count").
		Joins("INNER JOIN (?) AS billable ON cdrs.id = billable.id", billableSub).
		Joins("LEFT JOIN dns ON cdrs.clid = dns.number").
		Group("cdrs.clid, dns.cnam").
		Order("count DESC").
		Limit(10).
		Scan(&origins).Error; err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	var destinations []cdrStat
	if err := s.d.Raw().Model(&types.CDR{}).
		Select("cdrs.dnis AS value, dns.cnam AS cnam, COUNT(cdrs.id) AS count").
		Joins("INNER JOIN (?) AS billable ON cdrs.id = billable.id", billableSub).
		Joins("LEFT JOIN dns ON cdrs.dnis = dns.number").
		Group("cdrs.dnis, dns.cnam").
		Order("count DESC").
		Limit(10).
		Scan(&destinations).Error; err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ctx["origins"] = origins
	ctx["destinations"] = destinations

	// — minutes-based stats —
	var minuteOrigins []cdrStatMinutes
	if err := s.d.Raw().Model(&types.CDR{}).
		Select(`cdrs.clid AS value, dns.cnam AS cnam, SUM(strftime('%s', cdrs.end) - strftime('%s', cdrs.start)) AS total_seconds, MAX(strftime('%s', cdrs.end) - strftime('%s', cdrs.start)) AS longest_second`).
		Joins("INNER JOIN (?) AS billable ON cdrs.id = billable.id", billableSub).
		Joins("LEFT JOIN dns ON cdrs.clid = dns.number").
		Group("cdrs.clid, dns.cnam").
		Order("total_seconds DESC").
		Limit(10).
		Scan(&minuteOrigins).Error; err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	var minuteDestinations []cdrStatMinutes
	if err := s.d.Raw().Model(&types.CDR{}).
		Select(`cdrs.dnis AS value, dns.cnam AS cnam, SUM(strftime('%s', cdrs.end) - strftime('%s', cdrs.start)) AS total_seconds, MAX(strftime('%s', cdrs.end) - strftime('%s', cdrs.start)) AS longest_second`).
		Joins("INNER JOIN (?) AS billable ON cdrs.id = billable.id", billableSub).
		Joins("LEFT JOIN dns ON cdrs.dnis = dns.number").
		Group("cdrs.dnis, dns.cnam").
		Order("total_seconds DESC").
		Limit(10).
		Scan(&minuteDestinations).Error; err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ctx["minuteOrigins"] = minuteOrigins
	ctx["minuteDestinations"] = minuteDestinations

	type cdrLongestCall struct {
		CLID     string `gorm:"column:clid"`
		CLIDCNAM string `gorm:"column:clid_cnam"`
		DNIS     string `gorm:"column:dnis"`
		DNISCNAM string `gorm:"column:dnis_cnam"`
		Duration float64 `gorm:"column:duration"`
	}

	var longestCalls []cdrLongestCall
	if err := s.d.Raw().Model(&types.CDR{}).
		Select(`cdrs.clid, clid_dns.cnam AS clid_cnam, cdrs.dnis, dnis_dns.cnam AS dnis_cnam, strftime('%s', cdrs.end) - strftime('%s', cdrs.start) AS duration`).
		Joins("INNER JOIN (?) AS billable ON cdrs.id = billable.id", billableSub).
		Joins("LEFT JOIN dns AS clid_dns ON cdrs.clid = clid_dns.number").
		Joins("LEFT JOIN dns AS dnis_dns ON cdrs.dnis = dnis_dns.number").
		Order("duration DESC").
		Limit(25).
		Scan(&longestCalls).Error; err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	ctx["longestCalls"] = longestCalls

	// — revenue by LEC —
	type lecRevenue struct {
		LECName      string
		TotalRevenue int
	}

	lecs, err := s.d.LECList(r.Context(), nil)
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}

	lecRevenues := make([]lecRevenue, 0, len(lecs))
	for _, lec := range lecs {
		bp := billing.NewProcessor(billing.WithDatabase(s.d.(*db.DB)))
		if err := bp.Preload(r.Context(), lec); err != nil {
			slog.Error("Failed to preload fees for LEC", "lec", lec.Name, "error", err)
			continue
		}

		total := 0
		accounts, err := s.d.AccountList(r.Context(), nil)
		if err != nil {
			slog.Error("Failed to list accounts for LEC", "lec", lec.Name, "error", err)
			continue
		}
		for i := range accounts {
			account, err := s.d.AccountGet(r.Context(), &types.Account{ID: accounts[i].ID})
			if err != nil {
				slog.Error("Failed to hydrate account", "account", account.ID, "error", err)
				continue
			}
			bill, err := bp.BillAccount(r.Context(), account, lec)
			if err != nil {
				slog.Error("Failed to bill account", "account", account.ID, "error", err)
				continue
			}
			if bill.Cost() > 0 {
				total += bill.Cost()
			}
		}
		if total > 0 {
			lecRevenues = append(lecRevenues, lecRevenue{
				LECName:      lec.Name,
				TotalRevenue: total,
			})
		}
	}
	sort.Slice(lecRevenues, func(i, j int) bool {
		return lecRevenues[i].TotalRevenue > lecRevenues[j].TotalRevenue
	})

	ctx["lecRevenues"] = lecRevenues

	s.doTemplate(w, r, "views/work/statistics.p2", ctx)
}
