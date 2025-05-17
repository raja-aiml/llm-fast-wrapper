package logs

import "time"

// MemoryLogger stores logs in memory and is intended for testing.
type MemoryLogger struct{
    Entries []struct{
        Prompt string
        Token string
        Timestamp time.Time
    }
}

func NewMemoryLogger() Logger {
    return &MemoryLogger{}
}

func (m *MemoryLogger) LogPrompt(p, t string, ts time.Time) error {
    m.Entries = append(m.Entries, struct{
        Prompt string
        Token string
        Timestamp time.Time
    }{p, t, ts})
    return nil
}
