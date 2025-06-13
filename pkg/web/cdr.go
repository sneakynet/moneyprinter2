package web

import (
	"net/http"
	"encoding/json"

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

	json.NewEncoder(w).Encode(map[string]uint{"ID": id})
}
