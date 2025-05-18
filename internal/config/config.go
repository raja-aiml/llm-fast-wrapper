package config

import (
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const DefaultModel = "gpt-4"

type CLIConfig struct {
	Query       string
	Model       string
	Temperature float64
	Markdown    bool
	Stream      bool
	BaseURL     string
	Output      string // json, yaml, text, markdown
	LogFile     string // Optional log file path
}

func NewCLIConfig() *CLIConfig {
	return &CLIConfig{}
}

func NewClient(apiKey, baseURL string) openai.Client {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return openai.NewClient(opts...)
}
