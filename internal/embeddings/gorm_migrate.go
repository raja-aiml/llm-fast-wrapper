package embeddings

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations.sql
var migrationSQLTemplate string

// MigrateWithGORM applies SQL migrations to set up pgvector and embeddings table/index.
func MigrateWithGORM(dsn string, dimension int) error {
	// Open GORM DB connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("gorm open: %w", err)
	}
	// Format and execute embedded SQL
	rawSQL := fmt.Sprintf(migrationSQLTemplate, dimension)
	stmts := strings.Split(rawSQL, ";")
	for _, stmt := range stmts {
		s := strings.TrimSpace(stmt)
		if s == "" {
			continue
		}
		if err := db.Exec(s).Error; err != nil {
			return fmt.Errorf("execute migration stmt '%s': %w", s, err)
		}
	}
	return nil
}
