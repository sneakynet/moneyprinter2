package cmd

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/sneakynet/moneyprinter2/pkg/web"
	"github.com/sneakynet/moneyprinter2/pkg/db"
)

var (
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start up a moneyprinter server",
		Run:   serveCmdRun,
	}
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serveCmdRun(c *cobra.Command, args []string) {
	db, err := db.New()
	if err != nil {
		slog.Error("Error creating database", "error", err)
		return
	}

	dbPath := os.Getenv("MONEYPRINTER_DB")
	if dbPath == "" {
		dbPath = "moneyprinter.db"
	}

	if err := db.Connect(dbPath); err != nil {
		slog.Error("Error connecting to database", "error", err)
		return
	}

	if err := db.Migrate(); err != nil {
		slog.Error("Error migrating database", "error", err)
		return
	}

	s, err := web.New(web.WithDB(db))
	if err != nil {
		slog.Error("Error creating webserver", "error", err)
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := s.Serve(":8000"); err != nil && err != http.ErrServerClosed {
			slog.Error("Error initializing", "error", err)
			quit <- syscall.SIGINT
		}
	}()

	slog.Info("MoneyPrinter is ready to go BRRRRRRRR")
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		slog.Error("Error during shutdown", "error", err)
		return
	}
	slog.Info("Goodbye!")
}
