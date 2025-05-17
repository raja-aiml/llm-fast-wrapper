package logs

import "time"

type Logger interface {
	LogPrompt(prompt, token string, ts time.Time) error
}
