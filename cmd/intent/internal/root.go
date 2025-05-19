package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	dir, ext, dbDSN string
	dbDim           int
	threshold       float64
	seedOnly, useDB bool

	rootCmd = &cobra.Command{
		Use:   "intent [query]",
		Short: "Classify query intent via prompt strategy matching",
		Args:  cobra.MaximumNArgs(1),
		Run:   runIntentCommand,
	}
)

func Execute() {
	bindFlags(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("CLI error:", err)
		os.Exit(1)
	}
}
