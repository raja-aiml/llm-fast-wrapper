package embeddings

import (
	_ "embed"
	"fmt"

	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var migrateLogger *zap.SugaredLogger

func init() {
	migrateLogger = logging.InitLogger("logs/migrate.log")
	migrateLogger.Debug("Initialized migration logger")
}

//go:embed scripts/migrations.sql
var migrationSQLTemplate string

// MigrateWithGORM applies the SQL migration to PostgreSQL using GORM with dimension interpolation.
func MigrateWithGORM(dsn string, dimension int) error {
	migrateLogger.Infof("Starting migration with DSN: %s and dimension: %d", dsn, dimension)

	// Open GORM DB
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		migrateLogger.Errorf("Failed to open database via GORM: %v", err)
		return fmt.Errorf("gorm open: %w", err)
	}
	migrateLogger.Info("Successfully connected to PostgreSQL via GORM")

	// Prepare SQL with proper dimension
	rawSQL := fmt.Sprintf(migrationSQLTemplate, dimension)
	migrateLogger.Debugf("Compiled migration SQL with vector(%d)", dimension)

	// Execute migration
	if err := db.Exec(rawSQL).Error; err != nil {
		migrateLogger.Errorf("Migration execution failed: %v", err)
		return fmt.Errorf("execute migration scripts: %w", err)
	}

	migrateLogger.Info("Migration completed successfully")
	return nil
}
