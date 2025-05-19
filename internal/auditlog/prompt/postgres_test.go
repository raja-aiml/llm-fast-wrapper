package prompt_test

import (
	"testing"
	"time"

	"github.com/raja.aiml/llm-fast-wrapper/internal/auditlog/prompt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestPostgresLogger(t *testing.T) prompt.QueryableLogger {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}

	// Auto-migrate using same model
	if err := db.AutoMigrate(&prompt.PromptLogEntry{}); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	return &prompt.PostgresLogger{DB: db}
}

func TestPostgresLogger_LogPrompt(t *testing.T) {
	logger := setupTestPostgresLogger(t)
	ts := time.Now()

	err := logger.LogPrompt("prompt", "token", ts)
	if err != nil {
		t.Fatalf("LogPrompt failed: %v", err)
	}

	entries, err := logger.GetRecentLogs(10)
	if err != nil {
		t.Fatalf("GetRecentLogs failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Prompt != "prompt" || entries[0].Token != "token" {
		t.Errorf("unexpected entry: %+v", entries[0])
	}
}

func TestPostgresLogger_LogResponse(t *testing.T) {
	logger := setupTestPostgresLogger(t)
	ts := time.Now()

	err := logger.LogResponse("prompt2", "response2", "token2", ts)
	if err != nil {
		t.Fatalf("LogResponse failed: %v", err)
	}

	entries, err := logger.GetRecentLogs(10)
	if err != nil {
		t.Fatalf("GetRecentLogs failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	entry := entries[0]
	if entry.Prompt != "prompt2" || entry.Response != "response2" || entry.Token != "token2" {
		t.Errorf("unexpected entry: %+v", entry)
	}
}
