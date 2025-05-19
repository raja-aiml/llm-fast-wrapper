package prompt

import (
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresLogger stores prompt and response logs using GORM.
type PostgresLogger struct {
	DB *gorm.DB // Exported for test injection
}

// NewPostgresLogger connects using GORM and ensures the prompt_log table.
func NewPostgresLogger(dsn string) (Logger, error) {
	cfg := &gorm.Config{}

	if os.Getenv("GORM_LOG_LEVEL") == "silent" {
		cfg.Logger = logger.Discard
	}

	db, err := gorm.Open(postgres.Open(dsn), cfg)
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&PromptLogEntry{}); err != nil {
		return nil, err
	}
	return &PostgresLogger{DB: db}, nil
}

// NewPostgresLoggerFromDB is used for injecting a mock/test GORM DB.
func NewPostgresLoggerFromDB(db *gorm.DB) QueryableLogger {
	return &PostgresLogger{DB: db}
}

// LogPrompt inserts a prompt-only record.
func (l *PostgresLogger) LogPrompt(prompt, token string, ts time.Time) error {
	entry := &PromptLogEntry{
		Prompt:    prompt,
		Token:     token,
		Timestamp: ts,
	}
	return l.DB.Create(entry).Error
}

// LogResponse inserts a full prompt + response record.
func (l *PostgresLogger) LogResponse(prompt, response, token string, ts time.Time) error {
	entry := &PromptLogEntry{
		Prompt:    prompt,
		Response:  response,
		Token:     token,
		Timestamp: ts,
	}
	return l.DB.Create(entry).Error
}

// GetRecentLogs returns the most recent prompt-response entries.
func (l *PostgresLogger) GetRecentLogs(limit int) ([]PromptLogEntry, error) {
	var entries []PromptLogEntry
	if err := l.DB.Order("timestamp DESC").Limit(limit).Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}
