package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/raja.aiml/llm-fast-wrapper/internal/config"
)

// LogAudit appends the prompt/response to a structured JSONL audit file.
func LogAudit(prompt, response string, cfg *config.CLIConfig) {
	entry := map[string]any{
		"time":     time.Now().Format(time.RFC3339),
		"model":    cfg.Model,
		"prompt":   prompt,
		"response": response,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal audit entry: %v\n", err)
		return
	}

	f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open audit log file: %v\n", err)
		return
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write audit data: %v\n", err)
	}
	if _, err := f.Write([]byte("\n")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write newline to audit log: %v\n", err)
	}
}
