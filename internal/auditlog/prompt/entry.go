package prompt

import "time"

type PromptLogEntry struct {
	ID        uint      `gorm:"primaryKey"`
	Prompt    string    `gorm:"type:text"`
	Response  string    `gorm:"type:text"`
	Token     string    `gorm:"index"`
	Timestamp time.Time `gorm:"autoCreateTime"`
}

type QueryableLogger interface {
	Logger
	GetRecentLogs(limit int) ([]PromptLogEntry, error)
}
