package prompt_test

import (
	"testing"
	"time"

	"github.com/raja.aiml/llm-fast-wrapper/internal/auditlog/prompt"
)

func TestMemoryLogger_LogPrompt(t *testing.T) {
	logger := prompt.NewMemoryLogger()

	ts := time.Now()
	err := logger.LogPrompt("prompt1", "token1", ts)
	if err != nil {
		t.Fatalf("LogPrompt failed: %v", err)
	}

	mem, ok := logger.(*prompt.MemoryLogger)
	if !ok {
		t.Fatalf("expected MemoryLogger")
	}
	if got := len(mem.Prompts); got != 1 {
		t.Errorf("expected 1 prompt, got %d", got)
	}
	entry := mem.Prompts[0]
	if entry.Prompt != "prompt1" || entry.Token != "token1" || !entry.Timestamp.Equal(ts) {
		t.Errorf("incorrect prompt entry: %+v", entry)
	}
}

func TestMemoryLogger_LogResponse(t *testing.T) {
	logger := prompt.NewMemoryLogger()

	ts := time.Now()
	err := logger.LogResponse("prompt2", "response2", "token2", ts)
	if err != nil {
		t.Fatalf("LogResponse failed: %v", err)
	}

	mem, ok := logger.(*prompt.MemoryLogger)
	if !ok {
		t.Fatalf("expected MemoryLogger")
	}
	if got := len(mem.Responses); got != 1 {
		t.Errorf("expected 1 response, got %d", got)
	}
	entry := mem.Responses[0]
	if entry.Prompt != "prompt2" || entry.Response != "response2" || entry.Token != "token2" || !entry.Timestamp.Equal(ts) {
		t.Errorf("incorrect response entry: %+v", entry)
	}
}
