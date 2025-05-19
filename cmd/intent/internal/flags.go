package internal

import "github.com/spf13/cobra"

func bindFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&dir, "dir", "strategies", "Directory containing .md strategy files")
	cmd.Flags().StringVar(&ext, "ext", ".md", "File extension for strategy files")
	cmd.Flags().Float64Var(&threshold, "threshold", 0.5, "Minimum similarity score to accept")
	cmd.Flags().StringVar(&dbDSN, "db-dsn", "", "Postgres DSN for pgvector store (optional)")
	cmd.Flags().IntVar(&dbDim, "db-dim", 1536, "Expected vector dimension for pgvector")
	cmd.Flags().BoolVar(&seedOnly, "seed-only", false, "Seed strategies and exit")
	cmd.Flags().BoolVar(&useDB, "use-db", false, "Use database for strategy search")
}
