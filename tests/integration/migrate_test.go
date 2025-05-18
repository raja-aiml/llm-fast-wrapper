//go:build integration
// +build integration

package integration_test

import (
   "context"
   "os"
   "testing"

   "github.com/stretchr/testify/require"
   "github.com/raja.aiml/llm-fast-wrapper/internal/embeddings"
   "gorm.io/driver/postgres"
   "gorm.io/gorm"
   "gorm.io/gorm/logger"
)

// TestMigrateWithGORM applies the embedded SQL migrations and verifies
// that the embeddings table is created and operable.
func TestMigrateWithGORM(t *testing.T) {
   // Obtain DSN from environment or use default for local Docker setup
   dsn := os.Getenv("EMBEDDINGS_DSN")
   if dsn == "" {
       dsn = "host=localhost port=5432 user=llm password=llm dbname=llmlogs sslmode=disable"
   }
   // Open GORM connection in silent logger mode to avoid noisy errors when skipping
   db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
       Logger: logger.Default.LogMode(logger.Silent),
   })
   if err != nil {
       t.Skipf("Skipping migration test: cannot connect to Postgres at DSN %q: %v", dsn, err)
   }
   sqlDB, err := db.DB()
   if err != nil {
       t.Skipf("Skipping migration test: cannot obtain sql.DB: %v", err)
   }
   if err := sqlDB.Ping(); err != nil {
       t.Skipf("Skipping migration test: cannot ping database: %v", err)
   }

   // Perform migrations with a small test dimension
   const dim = 3
   require.NoError(t, embeddings.MigrateWithGORM(dsn, dim), "MigrateWithGORM should not return an error")

   // Confirm that the embeddings table exists
   exists := db.Migrator().HasTable("embeddings")
   require.True(t, exists, "embeddings table should exist after migration")

   // Test that the store can insert and retrieve data
   store, err := embeddings.NewPostgresStore(dsn, dim)
   require.NoError(t, err, "NewPostgresStore should succeed after migration")
   ctx := context.Background()
   sampleText := "migration-test-text"
   sampleVec := []float32{0, 1, 2}
   require.NoError(t, store.Store(ctx, sampleText, sampleVec), "Store should insert embedding successfully")
   got, err := store.Get(ctx, sampleText)
   require.NoError(t, err, "Get should retrieve embedding without error")
   require.Equal(t, sampleVec, got, "retrieved embedding should match stored value")
}