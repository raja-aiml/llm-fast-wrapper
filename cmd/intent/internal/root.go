package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfg = &Config{} // Global config instance passed to all modules

// rootCmd defines the CLI command
var rootCmd = &cobra.Command{
	Use:   "intent [query]",
	Short: "Classify query intent via prompt strategy matching",
	Args:  cobra.MaximumNArgs(1),
	Run:   runIntentCommand, // Matches required cobra.Command signature
}

// Execute sets up flags and runs the root command
func Execute() {
	BindFlags(rootCmd, cfg)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("CLI error:", err)
		os.Exit(1)
	}
}
