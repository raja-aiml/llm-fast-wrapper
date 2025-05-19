package prompt_test

import (
	"testing"
	"time"

	prompt "github.com/raja.aiml/llm-fast-wrapper/internal/auditlog/prompt"
)

func TestMemoryLogger_LogPrompt(t *testing.T) {
	logger := prompt.NewMemoryLogger()
	if err := logger.LogPrompt("prompt1", "token1", time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ml, ok := logger.(*prompt.MemoryLogger)
	if !ok {
		t.Fatalf("expected *MemoryLogger type")
	}

	if len(ml.Prompts) != 1 {
		t.Errorf("expected 1 prompt entry, got %d", len(ml.Prompts))
	}
}

func TestMemoryLogger_LogResponse(t *testing.T) {
	logger := prompt.NewMemoryLogger()
	if err := logger.LogResponse("prompt2", "response2", "token2", time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ml, ok := logger.(*prompt.MemoryLogger)
	if !ok {
		t.Fatalf("expected *MemoryLogger type")
	}

	if len(ml.Responses) != 1 {
		t.Errorf("expected 1 response entry, got %d", len(ml.Responses))
	}

	entry := ml.Responses[0]
	if entry.Prompt != "prompt2" || entry.Response != "response2" || entry.Token != "token2" {
		t.Errorf("unexpected response log entry: %+v", entry)
	}
}
