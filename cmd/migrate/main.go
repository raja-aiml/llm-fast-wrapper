package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/raja.aiml/llm-fast-wrapper/internal/embeddings/storage/postgres"
	"github.com/spf13/cobra"
)

var (
	dbDSN string
	dbDim int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "embeddings-cli",
		Short: "CLI for managing pgvector embeddings schema",
	}

	rootCmd.PersistentFlags().StringVar(&dbDSN, "db-dsn", "", "PostgreSQL DSN (required)")
	_ = rootCmd.MarkPersistentFlagRequired("db-dsn")

	rootCmd.AddCommand(MigrateCmd())
	rootCmd.AddCommand(DropCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Command failed: %v\n", err)
		os.Exit(1)
	}
}

// NewMigrateCmd returns the `migrate` CLI command
func MigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations for embeddings",
		Run: func(cmd *cobra.Command, args []string) {
			runMigrate()
		},
	}
	cmd.Flags().IntVar(&dbDim, "db-dim", 1536, "Embedding vector dimension (default 1536)")
	return cmd
}

// NewDropCmd returns the `drop` CLI command
func DropCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "drop",
		Short: "Drop the embeddings table (dev only!)",
		Run: func(cmd *cobra.Command, args []string) {
			runDrop()
		},
	}
}

// runMigrate applies the embedded migration script
func runMigrate() {
	fmt.Println("🔧 Running migration with dimension:", dbDim)
	if err := postgres.MigrateWithGORM(dbDSN, dbDim); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Migration completed successfully")
}

// runDrop explicitly drops the embeddings table
func runDrop() {
	fmt.Println("🗑️  Dropping table: embeddings ...")
	db, err := sql.Open("postgres", dbDSN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if _, err := db.Exec(`DROP TABLE IF EXISTS embeddings`); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Drop failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ embeddings table dropped successfully")
}

// go run cmd/migrate/main.go migrate --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable" --db-dim 1536
// go run cmd/migrate/main.go drop --db-dsn "postgresql://llm:llm@localhost:5432/llmlogs?sslmode=disable"
