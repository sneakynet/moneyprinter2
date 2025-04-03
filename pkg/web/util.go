package web

import (
	"encoding/csv"
	"io"
	"log/slog"
	"strconv"
	"strings"
)

func (s *Server) csvToMap(reader io.Reader) []map[string]string {
	r := csv.NewReader(reader)
	rows := []map[string]string{}
	var header []string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("Error decoding CSV", "error", err)
			return nil
		}
		if header == nil {
			header = record
		} else {
			dict := map[string]string{}
			for i := range header {
				dict[header[i]] = strings.TrimSpace(record[i])
			}
			rows = append(rows, dict)
		}
	}
	return rows
}

func (s *Server) strToUint(st string) uint {
	int, err := strconv.Atoi(st)
	if err != nil {
		return 0
	}
	return uint(int)
}
