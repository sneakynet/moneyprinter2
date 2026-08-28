package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/sneakynet/moneyprinter2/pkg/cdr"
)

var (
	cdrFollow bool

	cdrCmd = &cobra.Command{
		Use:   "cdr <CLLI> <type> <file>",
		Short: "Ingest CDRs from file",
		Run:   cdrCmdRun,
		Args:  cobra.ExactArgs(3),
	}
)

func init() {
	rootCmd.AddCommand(cdrCmd)
	cdrCmd.Flags().BoolVarP(&cdrFollow, "follow", "f", false, "Continuously parse and upload new CDRs as they become available")
}

func cdrCmdRun(c *cobra.Command, args []string) {
	_, skipUpload := os.LookupEnv("SKIP_UPLOAD")
	mpAddr := os.Getenv("MP_ADDR")
	if mpAddr == "" {
		mpAddr = "localhost:8000"
	}

	var parser cdr.Parser
	switch args[1] {
	case "cisco":
		parser = new(cdr.Cisco)
	case "meridian":
		parser = new(cdr.Meridian)
	default:
		slog.Error("Invalid CDR type; valid options are 'cisco' or 'meridian'")
		return
	}

	if !cdrFollow {
		cdrCmdBatch(args[0], args[2], parser, skipUpload, mpAddr)
		return
	}

	cdrCmdFollow(args[0], args[2], parser, skipUpload, mpAddr)
}

func cdrCmdBatch(clli, file string, parser cdr.Parser, skipUpload bool, mpAddr string) {
	f, err := os.Open(file)
	if err != nil {
		slog.Error("Error loading CDR", "error", err)
		return
	}
	defer f.Close()

	records, err := parser.Parse(f, clli)
	if err != nil {
		slog.Error("Error parsing CDRs", "error", err)
		return
	}

	for i, r := range records {
		slog.Info("record", "number", i, "from", r.CLID, "to", r.DNIS, "duration", r.End.Sub(r.Start))
		if !skipUpload {
			uploadCDR(r, mpAddr)
		}
	}
}

func cdrCmdFollow(clli, file string, parser cdr.Parser, skipUpload bool, mpAddr string) {
	f, err := os.Open(file)
	if err != nil {
		slog.Error("Error loading CDR", "error", err)
		return
	}
	defer f.Close()

	// Parse initial content
	records, err := parser.Parse(f, clli)
	if err != nil {
		slog.Error("Error parsing CDRs", "error", err)
		return
	}

	recordNum := 0
	for _, r := range records {
		slog.Info("record", "number", recordNum, "from", r.CLID, "to", r.DNIS, "duration", r.End.Sub(r.Start))
		if !skipUpload {
			uploadCDR(r, mpAddr)
		}
		recordNum++
	}

	slog.Info("Following file for new CDRs", "file", file)

	// Track file position and buffer incomplete trailing data.
	offset, _ := f.Seek(0, io.SeekCurrent)
	incomplete := []byte{}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		stat, err := f.Stat()
		if err != nil {
			slog.Error("Error stating file", "error", err)
			return
		}

		// No new data to read
		if stat.Size() <= offset {
			continue
		}

		// Read new bytes from the tracked offset
		if _, err := f.Seek(offset, 0); err != nil {
			slog.Error("Error seeking file", "error", err)
			return
		}

		newData, err := io.ReadAll(io.LimitReader(f, stat.Size()-offset))
		if err != nil {
			slog.Error("Error reading new data", "error", err)
			return
		}

		// Prepend any incomplete data from previous iteration
		data := append(incomplete, newData...)
		incomplete = nil

		// Find the last newline to separate complete lines from trailing partial data
		lastNewline := bytes.LastIndexByte(data, '\n')
		if lastNewline < 0 {
			// No complete lines yet — buffer everything for next round
			incomplete = data
			continue
		}

		complete := data[:lastNewline+1]
		if lastNewline+1 < int(len(data)) {
			incomplete = data[lastNewline+1:]
		}

		records, err := parser.Parse(bytes.NewReader(complete), clli)
		if err != nil {
			slog.Error("Error parsing new CDRs", "error", err)
			offset += int64(len(data))
			continue
		}

		for _, r := range records {
			slog.Info("record", "number", recordNum, "from", r.CLID, "to", r.DNIS, "duration", r.End.Sub(r.Start))
			if !skipUpload {
				uploadCDR(r, mpAddr)
			}
			recordNum++
		}

		offset += int64(len(data))
	}
}

func uploadCDR(r any, mpAddr string) {
	cl := http.Client{Timeout: time.Second * 10}

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(r)
	req, _ := http.NewRequest(http.MethodPost, "http://"+mpAddr+"/api/admin/usage/cdr/ingest", buf)
	req.SetBasicAuth(os.Getenv("MP_USERNAME"), os.Getenv("MP_PASSWORD"))
	resp, err := cl.Do(req)
	if err != nil {
		slog.Error("Error uploading CDR", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		msg := make(map[string]string)
		if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
			slog.Error("Error encountered while decoding response", "error", err)
			return
		}

		slog.Warn("Insert refused", "message", msg)
		return
	}
	res := make(map[string]uint)
	json.NewDecoder(resp.Body).Decode(&res)
	slog.Info("Created new CDR", "ID", res["ID"])
}
