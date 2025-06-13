package web

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/flosch/pongo2/v6"

	"github.com/sneakynet/moneyprinter2/pkg/types"
)

func (s *Server) apiCDRIngest(w http.ResponseWriter, r *http.Request) {
	cdr := new(types.CDR)

	if err := json.NewDecoder(r.Body).Decode(cdr); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	id, err := s.d.CDRSave(cdr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]uint{"ID": id})
}

func (s *Server) uiViewCDRList(w http.ResponseWriter, r *http.Request) {
	cdrs := []types.CDR{}

	q := r.URL.Query()

	if dn := q.Get("dn"); dn != "" {
		q.Set("dnis", dn)
		q.Set("clid", dn)
	}

	if dnis := q.Get("dnis"); dnis != "" {
		s1, _ := s.d.CDRList(&types.CDR{DNIS: dnis})
		cdrs = append(cdrs, s1...)
	}

	if clid := q.Get("clid"); clid != "" {
		s1, _ := s.d.CDRList(&types.CDR{CLID: clid})
		cdrs = append(cdrs, s1...)
	}

	if clli := q.Get("ccli"); clli != "" {
		s1, _ := s.d.CDRList(&types.CDR{CLLI: clli})
		cdrs = append(cdrs, s1...)
	}

	sort.Slice(cdrs, func(i, j int) bool {
		return cdrs[i].Start.Before(cdrs[j].Start)
	})

	s.doTemplate(w, r, "views/cdr/list.p2", pongo2.Context{"cdrs": cdrs})
}
