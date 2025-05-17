package config

import (
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

const DefaultModel = "gpt-4-1106-preview"

type CLIConfig struct {
	Query       string
	Model       string
	Temperature float64
	Markdown    bool
	Stream      bool
	BaseURL     string
}

func NewCLIConfig() *CLIConfig {
	return &CLIConfig{}
}

func NewClient(apiKey string, baseURL string) openai.Client {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return openai.NewClient(opts...)
}
