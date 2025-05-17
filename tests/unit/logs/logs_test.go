package logs_test

import (
    "testing"
    "time"

    "github.com/your-org/llm-fast-wrapper/internal/logs"
)

func TestMemoryLogger(t *testing.T) {
    logger := logs.NewMemoryLogger()
    if err := logger.LogPrompt("p", "t", time.Now()); err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    ml, ok := logger.(*logs.MemoryLogger)
    if !ok {
        t.Fatalf("expected MemoryLogger type")
    }
    if len(ml.Entries) != 1 {
        t.Fatalf("expected 1 entry, got %d", len(ml.Entries))
    }
}
