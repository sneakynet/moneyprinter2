package cmd

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "moneyprinter",
		Short: "Entrypoint for all MoneyPrinter commands",
		Long:  rootCmdLongDocs,
	}

	rootCmdLongDocs = `MoneyPrinter is a complete telephone management and billing suite, capable of managing logistics for a moderate sized show telephone network.`
)

// loadEnvFile reads KEY=VALUE pairs from the given file and sets them
// as environment variables. Lines starting with # or blank lines are
// skipped.
func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}
	return s.Err()
}

// Entrypoint is the root node on the command tree.
func Entrypoint() {
	// If invoked as mp2-bill-shell, load environment from the config
	// file before reading any env vars, then prepend "billui" so cobra
	// dispatches directly to the billing TUI (busybox-style alias).
	if base := path.Base(os.Args[0]); base == "mp2-bill-shell" {
		if err := loadEnvFile("/etc/moneyprinter2/billing.env"); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "warning: failed to load billing env: %v\n", err)
		}
		os.Args = append([]string{os.Args[0], "billui"}, os.Args[1:]...)
	}

	logLevel := os.Getenv("LOG_LEVEL")
	var level slog.Level
	switch strings.ToLower(logLevel) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr,
		&slog.HandlerOptions{Level: level}))
	slog.SetDefault(logger)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}
