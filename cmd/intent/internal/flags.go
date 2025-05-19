package internal

import "github.com/spf13/cobra"

// BindFlags connects CLI flags to the Config struct
func BindFlags(cmd *cobra.Command, cfg *Config) {
	cmd.Flags().StringVar(&cfg.Dir, "dir", "strategies", "Directory containing .md strategy files")
	cmd.Flags().StringVar(&cfg.Ext, "ext", ".md", "File extension for strategy files")
	cmd.Flags().Float64Var(&cfg.Threshold, "threshold", 0.5, "Minimum similarity score to accept")

	cmd.Flags().StringVar(&cfg.DbDSN, "db-dsn", "", "Postgres DSN for pgvector store (optional)")
	cmd.Flags().IntVar(&cfg.DbDim, "db-dim", 1536, "Expected vector dimension for pgvector")

	cmd.Flags().BoolVar(&cfg.SeedOnly, "seed-only", false, "Seed strategies and exit")
	cmd.Flags().BoolVar(&cfg.UseDB, "use-db", false, "Use database for strategy search")
}
