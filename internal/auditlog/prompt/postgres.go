package prompt

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresLogger stores prompt and response logs using GORM.
type PostgresLogger struct {
	db *gorm.DB
}

// NewPostgresLogger connects using GORM and ensures the llm_logs table.
func NewPostgresLogger(dsn string) (Logger, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the llm_logs table
	if err := db.AutoMigrate(&LLMLog{}); err != nil {
		return nil, err
	}

	return &PostgresLogger{db: db}, nil
}

// LogPrompt inserts a prompt-only record.
func (l *PostgresLogger) LogPrompt(prompt, token string, ts time.Time) error {
	entry := &LLMLog{
		Prompt:    prompt,
		Token:     token,
		Timestamp: ts,
	}
	return l.db.Create(entry).Error
}

// LogResponse inserts a full prompt + response record.
func (l *PostgresLogger) LogResponse(prompt, response, token string, ts time.Time) error {
	entry := &LLMLog{
		Prompt:    prompt,
		Response:  response,
		Token:     token,
		Timestamp: ts,
	}
	return l.db.Create(entry).Error
}

// GetRecentLogs returns the most recent prompt-response entries (ordered by timestamp DESC).
func (l *PostgresLogger) GetRecentLogs(limit int) ([]LLMLog, error) {
	var entries []LLMLog
	if err := l.db.Order("timestamp DESC").Limit(limit).Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}
