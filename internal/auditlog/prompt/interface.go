package prompt

import "time"

// Logger defines methods for logging prompts and responses.
type Logger interface {
	LogPrompt(prompt, token string, ts time.Time) error
	LogResponse(prompt, response, token string, ts time.Time) error
}
