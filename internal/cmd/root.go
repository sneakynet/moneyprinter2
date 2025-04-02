package cmd

import (
	"fmt"
	"os"

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

// Entrypoint is the root node on the command tree.
func Entrypoint() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}
