package cmd

import (
	"log/slog"
	"os"
	"strings"

	"github.com/sneakynet/moneyprinter2/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var billuiCmd = &cobra.Command{
	Use:   "billui",
	Short: "Interactive bill printer UI",
	Run:   billuiCmdRun,
}

func init() {
	rootCmd.AddCommand(billuiCmd)
}

func billuiCmdRun(_ *cobra.Command, _ []string) {
	client := tui.NewClient()
	printers := tui.ParsePrinters(os.Getenv("MONEYPRINTER_PRINTERS"))
	greeting := strings.ReplaceAll(os.Getenv("MONEYPRINTER_GREETING"), "\\n", "\n")
	m := tui.NewBillViewer(client, greeting, printers)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error("TUI program failed", "error", err)
		os.Exit(1)
	}
}
