package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/raja.aiml/llm-fast-wrapper/client/internal/chat"
	"github.com/raja.aiml/llm-fast-wrapper/client/internal/help"
	"github.com/raja.aiml/llm-fast-wrapper/internal/config"
	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
)

func main() {
	_ = godotenv.Load()
	logger := logging.InitLogger()
	cfg := config.NewCLIConfig()

	rootCmd := &cobra.Command{
		Use:     "llm-client",
		Short:   "CLI to interact with OpenAI-compatible LLM",
		Example: help.LoadUsageMarkdown("client/internal/help/usage.md"),
		Run: func(cmd *cobra.Command, args []string) {
			apiKey := os.Getenv("OPENAI_API_KEY")
			if apiKey == "" {
				logger.Fatal("Missing environment variable: OPENAI_API_KEY")
			}

			var opts []option.RequestOption
			opts = append(opts, option.WithAPIKey(apiKey))
			if cfg.BaseURL != "" {
				opts = append(opts, option.WithBaseURL(cfg.BaseURL))
			}

			client := openai.NewClient(opts...)

			if cfg.Query != "" {
				chat.RunQuery(&client, cfg, logger)
			} else {
				chat.RunInteractive(&client, cfg, logger)
			}
		},
	}

	rootCmd.Flags().StringVarP(&cfg.Query, "query", "q", "", "Prompt to send to the model")
	rootCmd.Flags().StringVarP(&cfg.Model, "model", "m", config.DefaultModel, "Model name")
	rootCmd.Flags().Float64VarP(&cfg.Temperature, "temperature", "t", 0.7, "Sampling temperature")
	rootCmd.Flags().BoolVar(&cfg.Stream, "stream", false, "Enable streaming response (set true to stream)")
	rootCmd.Flags().StringVar(&cfg.BaseURL, "base-url", "", "Custom OpenAI-compatible base URL")
	rootCmd.Flags().StringVar(&cfg.Output, "output", "text", "Output format: text, markdown, json, yaml")
	rootCmd.Flags().StringVar(&cfg.LogFile, "log-file", "llm-client.log", "Path to log file for prompts/responses")
	rootCmd.Flags().BoolVar(&cfg.AuditEnabled, "audit", true, "Enable audit logging of prompt/response")

	rootCmd.AddCommand(helpCommand())

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("Command execution failed: %v", err)
	}
}

func helpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "Show full Markdown-based usage guide",
		Run: func(cmd *cobra.Command, args []string) {
			md := help.LoadUsageMarkdown("client/help/usage.md")
			fmt.Println(md)
		},
	}
}
