package prompt_test

import (
	"os"
	"testing"
	"time"

	"github.com/raja.aiml/llm-fast-wrapper/internal/auditlog/prompt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewPostgresLoggerFromDB(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}

	logger := prompt.NewPostgresLoggerFromDB(db)
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNewPostgresLogger_InvalidDSN(t *testing.T) {
	// Suppress GORM internal error output from polluting test logs
	_ = os.Setenv("GORM_LOG_LEVEL", "silent")

	_, err := prompt.NewPostgresLogger("invalid-dsn-should-fail")
	if err == nil {
		t.Fatal("expected error for invalid DSN, got nil")
	}
}

func TestNewPostgresLogger_ValidDSN(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	// Manually wrap and test
	logger := &prompt.PostgresLogger{DB: db}
	if err := db.AutoMigrate(&prompt.PromptLogEntry{}); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	ts := time.Now()
	if err := logger.LogPrompt("Hello", "token", ts); err != nil {
		t.Fatalf("LogPrompt failed: %v", err)
	}

	logs, err := logger.GetRecentLogs(1)
	if err != nil {
		t.Fatalf("GetRecentLogs failed: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(logs))
	}
}
