package prompt

import "time"

// MemoryLogger stores prompt and response logs in memory.
type MemoryLogger struct {
	Prompts   []PromptEntry
	Responses []ResponseEntry
}

type PromptEntry struct {
	Prompt    string
	Token     string
	Timestamp time.Time
}

type ResponseEntry struct {
	Prompt    string
	Response  string
	Token     string
	Timestamp time.Time
}

// NewMemoryLogger creates a new in-memory logger instance.
func NewMemoryLogger() Logger {
	return &MemoryLogger{}
}

// LogPrompt stores a prompt entry in memory.
func (m *MemoryLogger) LogPrompt(p, t string, ts time.Time) error {
	m.Prompts = append(m.Prompts, PromptEntry{p, t, ts})
	return nil
}

// LogResponse stores a prompt + response entry in memory.
func (m *MemoryLogger) LogResponse(p, r, t string, ts time.Time) error {
	m.Responses = append(m.Responses, ResponseEntry{p, r, t, ts})
	return nil
}
