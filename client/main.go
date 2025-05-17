package main

import (
	"os"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/raja.aiml/llm-fast-wrapper/client/internal/chat"
	"github.com/raja.aiml/llm-fast-wrapper/internal/config"
	"github.com/raja.aiml/llm-fast-wrapper/internal/logging"
)

func main() {
	_ = godotenv.Load()
	logger := logging.InitLogger()
	cfg := config.NewCLIConfig()

	rootCmd := &cobra.Command{
		Use:   "llm-client",
		Short: "CLI to interact with OpenAI-compatible LLM",
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
	rootCmd.Flags().BoolVar(&cfg.Markdown, "markdown", false, "Render output in Markdown")
	rootCmd.Flags().BoolVar(&cfg.Stream, "stream", true, "Enable streaming response")
	rootCmd.Flags().StringVar(&cfg.BaseURL, "base-url", "", "Custom OpenAI-compatible base URL")

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("Command execution failed: %v", err)
	}
}
