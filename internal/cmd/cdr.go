package cmd

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/sneakynet/moneyprinter2/pkg/cdr"
)

var (
	cdrCmd = &cobra.Command{
		Use:   "cdr <CLLI> <type> <file>",
		Short: "Ingest CDRs from file",
		Run:   cdrCmdRun,
		Args:  cobra.ExactArgs(3),
	}
)

func init() {
	rootCmd.AddCommand(cdrCmd)
}

func cdrCmdRun(c *cobra.Command, args []string) {
	_, skipUpload := os.LookupEnv("SKIP_UPLOAD")
	mpAddr := os.Getenv("MP_ADDR")
	if mpAddr == "" {
		mpAddr = "localhost:8000"
	}

	f, err := os.Open(args[2])
	if err != nil {
		slog.Error("Error loading CDR", "error", err)
		return
	}
	defer f.Close()

	var parser cdr.Parser

	switch args[1] {
	case "cisco":
		parser = new(cdr.Cisco)
	default:
		slog.Error("Invalid CDR type; valid options are 'cisco'")
		return
	}

	records, err := parser.Parse(f, args[0])
	if err != nil {
		slog.Error("Error parsing CDRs")
		return
	}

	for i, r := range records {
		slog.Info("record", "number", i, "from", r.CLID, "to", r.DNIS, "duration", r.End.Sub(r.Start))

		if !skipUpload {
			buf := new(bytes.Buffer)
			json.NewEncoder(buf).Encode(r)
			resp, err := http.Post("http://"+mpAddr+"/api/admin/usage/cdr/ingest", "application/json", buf)
			if err != nil {
				slog.Error("Error uploading CDR", "error", err)
				continue
			}
			if resp.StatusCode != http.StatusCreated {
				msg := make(map[string]string)
				if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
					slog.Error("Error encountered while decoding response", "error", err)
					continue
				}

				slog.Warn("Insert refused", "message", msg)
				continue
			}
			res := make(map[string]uint)
			json.NewDecoder(resp.Body).Decode(&res)
			slog.Info("Created new CDR", "ID", res["ID"])
		}
	}
}
