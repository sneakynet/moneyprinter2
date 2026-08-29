package cmd

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	billCmd = &cobra.Command{
		Use:   "bill <id>",
		Short: "bill an account",
		Run:   billCmdRun,
		Args:  cobra.ExactArgs(1),
	}
)

func init() {
	rootCmd.AddCommand(billCmd)
}

func billCmdRun(c *cobra.Command, args []string) {
	mpAddr := os.Getenv("MP_ADDR")
	if mpAddr == "" {
		mpAddr = "localhost:8000"
	}

	cl := http.Client{Timeout: time.Second * 10}

	req := &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Scheme: "http",
			Host:   mpAddr,
			Path:   fmt.Sprintf("/api/admin/money/bills/by-account/%s?lec=1", args[0]),
			User: url.UserPassword(
				os.Getenv("MP_USERNAME"),
				os.Getenv("MP_PASSWORD"),
			),
		},
	}
	resp, err := cl.Do(req)
	if err != nil {
		slog.Error("Error billing account", "error", err)
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading response", "error", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		slog.Error("Bill request failed", "status", resp.StatusCode, "body", string(b))
		return
	}
	fmt.Print(string(b))
}
