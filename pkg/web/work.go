package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	"github.com/flosch/pongo2/v6"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/leekchan/accounting"

	"github.com/sneakynet/moneyprinter2/pkg/billing"
	"github.com/sneakynet/moneyprinter2/pkg/db"
	"github.com/sneakynet/moneyprinter2/pkg/types"
)

type serviceCount struct {
	Service string
	Slug    string
	Count   int
}

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

type cdrLongestCall struct {
	CLID     string  `gorm:"column:clid"`
	CLIDCNAM string  `gorm:"column:clid_cnam"`
	DNIS     string  `gorm:"column:dnis"`
	DNISCNAM string  `gorm:"column:dnis_cnam"`
	Duration float64 `gorm:"column:duration"`
}

type lecRevenue struct {
	LECName      string
	TotalRevenue int
}

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

func (s *Server) getStatisticsContext(rctx context.Context) (pongo2.Context, error) {
	ctx := pongo2.Context{}

	svcs, err := s.d.ServiceList(rctx, &types.Service{})
	if err != nil {
		return nil, err
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
		return nil, err
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
		return nil, err
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
		return nil, err
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
		return nil, err
	}

	ctx["minuteOrigins"] = minuteOrigins
	ctx["minuteDestinations"] = minuteDestinations

	var longestCalls []cdrLongestCall
	if err := s.d.Raw().Model(&types.CDR{}).
		Select(`cdrs.clid, clid_dns.cnam AS clid_cnam, cdrs.dnis, dnis_dns.cnam AS dnis_cnam, strftime('%s', cdrs.end) - strftime('%s', cdrs.start) AS duration`).
		Joins("INNER JOIN (?) AS billable ON cdrs.id = billable.id", billableSub).
		Joins("LEFT JOIN dns AS clid_dns ON cdrs.clid = clid_dns.number").
		Joins("LEFT JOIN dns AS dnis_dns ON cdrs.dnis = dnis_dns.number").
		Order("duration DESC").
		Limit(25).
		Scan(&longestCalls).Error; err != nil {
		return nil, err
	}

	ctx["longestCalls"] = longestCalls

	lecs, err := s.d.LECList(rctx, nil)
	if err != nil {
		return nil, err
	}

	lecRevenues := make([]lecRevenue, 0, len(lecs))
	for _, lec := range lecs {
		bp := billing.NewProcessor(billing.WithDatabase(s.d.(*db.DB)))
		if err := bp.Preload(rctx, lec); err != nil {
			slog.Error("Failed to preload fees for LEC", "lec", lec.Name, "error", err)
			continue
		}

		total := 0
		accounts, err := s.d.AccountList(rctx, nil)
		if err != nil {
			slog.Error("Failed to list accounts for LEC", "lec", lec.Name, "error", err)
			continue
		}
		for i := range accounts {
			account, err := s.d.AccountGet(rctx, &types.Account{ID: accounts[i].ID})
			if err != nil {
				slog.Error("Failed to hydrate account", "account", account.ID, "error", err)
				continue
			}
			bill, err := bp.BillAccount(rctx, account, lec)
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
	return ctx, nil
}

func (s *Server) uiViewWorkStatistics(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getStatisticsContext(r.Context())
	if err != nil {
		s.doTemplate(w, r, "errors/internal.p2", pongo2.Context{"error": err.Error()})
		return
	}
	s.doTemplate(w, r, "views/work/statistics.p2", ctx)
}

func (s *Server) apiWorkStatistics(w http.ResponseWriter, r *http.Request) {
	ctx, err := s.getStatisticsContext(r.Context())
	if err != nil {
		json.NewEncoder(w).Encode(pongo2.Context{"error": err})
		return
	}

	// Process CTX here
	width := s.strToUint(r.URL.Query().Get("width"))
	if width == 0 {
		width = 80
	}

	ac := accounting.Accounting{Symbol: "$", Precision: 2}

	formatDur := func(seconds float64) string {
		h := int(seconds) / 3600
		m := int(seconds) % 3600 / 60
		s := int(seconds) % 60
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}

	services, _ := ctx["services"].([]serviceCount)
	origins, _ := ctx["origins"].([]cdrStat)
	destinations, _ := ctx["destinations"].([]cdrStat)
	minuteOrigins, _ := ctx["minuteOrigins"].([]cdrStatMinutes)
	minuteDestinations, _ := ctx["minuteDestinations"].([]cdrStatMinutes)
	longestCalls, _ := ctx["longestCalls"].([]cdrLongestCall)
	lecRevenues, _ := ctx["lecRevenues"].([]lecRevenue)

	rowConfig := table.RowConfig{
		AutoMerge:      true,
		AutoMergeAlign: text.AlignCenter, // Center-align the merged text
	}

	// Table 1: Service Consumption
	t := table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Size.WidthMin = int(width - 3)
	t.SetOutputMirror(w)
	t.SetAllowedRowLength(int(width))
	t.AppendHeader(table.Row{
		"Service Consumption",
		"Service Consumption",
		"Service Consumption",
	}, rowConfig)
	t.AppendHeader(table.Row{"Service", "Slug", "Count"})
	if len(services) == 0 {
		t.AppendRow(table.Row{"No data found"})
	} else {
		for _, svc := range services {
			t.AppendRow(table.Row{svc.Service, svc.Slug, svc.Count})
		}
	}
	t.Render()
	fmt.Fprintln(w, "")

	// Table 2: Top 10 Origins by Call Volume
	t = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Size.WidthMin = int(width - 3)
	t.SetOutputMirror(w)
	t.SetAllowedRowLength(int(width))
	t.AppendHeader(table.Row{
		"Top 10 Origins by Call Volume",
		"Top 10 Origins by Call Volume",
		"Top 10 Origins by Call Volume",
	}, rowConfig)
	t.AppendHeader(table.Row{"CLID", "CNAM", "Call Count"})
	if len(origins) == 0 {
		t.AppendRow(table.Row{"No data found"})
	} else {
		for _, o := range origins {
			t.AppendRow(table.Row{o.Value, o.CNAM, o.Count})
		}
	}
	t.Render()
	fmt.Fprintln(w, "")

	// Table 3: Top 10 Destinations by Call Volume
	t = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Size.WidthMin = int(width - 3)
	t.SetOutputMirror(w)
	t.SetAllowedRowLength(int(width))
	t.AppendHeader(table.Row{
		"Top 10 Destinations by Call Volume",
		"Top 10 Destinations by Call Volume",
		"Top 10 Destinations by Call Volume",
	}, rowConfig)
	t.AppendHeader(table.Row{"DNIS", "CNAM", "Call Count"})
	if len(destinations) == 0 {
		t.AppendRow(table.Row{"No data found"})
	} else {
		for _, d := range destinations {
			t.AppendRow(table.Row{d.Value, d.CNAM, d.Count})
		}
	}
	t.Render()
	fmt.Fprintln(w, "")

	// Table 4: Top 10 Origins by Billable Call Minutes
	t = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Size.WidthMin = int(width - 3)
	t.SetOutputMirror(w)
	t.SetAllowedRowLength(int(width))
	t.AppendHeader(table.Row{
		"Top 10 Origins by Billable Call Minutes",
		"Top 10 Origins by Billable Call Minutes",
		"Top 10 Origins by Billable Call Minutes",
		"Top 10 Origins by Billable Call Minutes",
	}, rowConfig)
	t.AppendHeader(table.Row{"DN", "CNAM", "Total Minutes", "Longest Call"})
	if len(minuteOrigins) == 0 {
		t.AppendRow(table.Row{"No data found"})
	} else {
		for _, o := range minuteOrigins {
			t.AppendRow(table.Row{o.Value, o.CNAM, formatDur(o.TotalSeconds), formatDur(o.LongestSecond)})
		}
	}
	t.Render()
	fmt.Fprintln(w, "")

	// Table 5: Top 10 Destinations by Billable Call Minutes
	t = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Size.WidthMin = int(width - 3)
	t.SetOutputMirror(w)
	t.SetAllowedRowLength(int(width))
	t.AppendHeader(table.Row{
		"Top 10 Destinations by Billable Call Minutes",
		"Top 10 Destinations by Billable Call Minutes",
		"Top 10 Destinations by Billable Call Minutes",
		"Top 10 Destinations by Billable Call Minutes",
	}, rowConfig)
	t.AppendHeader(table.Row{"DN", "CNAM", "Total Minutes", "Longest Call"})
	if len(minuteDestinations) == 0 {
		t.AppendRow(table.Row{"No data found"})
	} else {
		for _, d := range minuteDestinations {
			t.AppendRow(table.Row{d.Value, d.CNAM, formatDur(d.TotalSeconds), formatDur(d.LongestSecond)})
		}
	}
	t.Render()
	fmt.Fprintln(w, "")

	// Table 6: Top 25 Longest Billable Calls
	t = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Size.WidthMin = int(width - 3)
	t.SetOutputMirror(w)
	t.SetAllowedRowLength(int(width))
	t.AppendHeader(table.Row{
		"Top 25 Longest Billable Calls",
		"Top 25 Longest Billable Calls",
		"Top 25 Longest Billable Calls",
		"Top 25 Longest Billable Calls",
		"Top 25 Longest Billable Calls",
	}, rowConfig)
	t.AppendHeader(table.Row{"Origin", "Origin CNAM", "Destination", "Destination CNAM", "Duration"})
	if len(longestCalls) == 0 {
		t.AppendRow(table.Row{"No data found"})
	} else {
		for _, c := range longestCalls {
			t.AppendRow(table.Row{c.CLID, c.CLIDCNAM, c.DNIS, c.DNISCNAM, formatDur(c.Duration)})
		}
	}
	t.Render()
	fmt.Fprintln(w, "")

	// Table 7: Total Revenue by LEC
	t = table.NewWriter()
	t.SetColumnConfigs([]table.ColumnConfig{{
		Name:        "Total Revenue",
		Align:       text.AlignRight,
		AlignHeader: text.AlignRight,
	}})
	t.Style().Options.DrawBorder = false
	t.Style().Size.WidthMin = int(width - 3)
	t.SetOutputMirror(w)
	t.SetAllowedRowLength(int(width))
	t.AppendHeader(table.Row{"Total Revenue by LEC", "Total Revenue by LEC"}, rowConfig)
	t.AppendHeader(table.Row{"LEC", "Total Revenue"})
	if len(lecRevenues) == 0 {
		t.AppendRow(table.Row{"No data found"})
	} else {
		for _, r := range lecRevenues {
			t.AppendRow(table.Row{r.LECName, ac.FormatMoney(float64(r.TotalRevenue) / 100)})
		}
	}
	t.Render()
}
