//go:build integration

package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "llm-fast-wrapper",
	Short: "LLM streaming server",
}

func Execute() error { return rootCmd.Execute() }

func init() {
	rootCmd.AddCommand(serveCmd)
}
