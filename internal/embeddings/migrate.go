package embeddings

import (
	_ "embed" // for embedding SQL
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed scripts/migrations.sql
var migrationSQLTemplate string

// MigrateWithGORM applies SQL migrations with a specified vector dimension.
func MigrateWithGORM(dsn string, dimension int) error {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("gorm open: %w", err)
	}

	// Replace dimension placeholder
	rawSQL := fmt.Sprintf(migrationSQLTemplate, dimension)

	if err := db.Exec(rawSQL).Error; err != nil {
		return fmt.Errorf("execute migration scripts: %w", err)
	}
	return nil
}
