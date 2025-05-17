package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	query       string
	model       string
	temperature float64
	markdown    bool
	stream      bool
	baseURL     string
	logger      *zap.SugaredLogger
)

func init() {
	_ = godotenv.Load()

	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"logs/llm-client.log"}
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logr, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	logger = logr.Sugar()
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "llm-client",
		Short: "CLI to interact with OpenAI-compatible LLM",
		Run: func(cmd *cobra.Command, args []string) {
			apiKey := os.Getenv("OPENAI_API_KEY")
			if apiKey == "" {
				logger.Fatal("Missing environment variable: OPENAI_API_KEY")
			}

			opts := []option.RequestOption{option.WithAPIKey(apiKey)}
			if baseURL != "" {
				opts = append(opts, option.WithBaseURL(baseURL))
			}

			client := openai.NewClient(opts...)
			logger.Infof("Connected to OpenAI-compatible endpoint")

			if query != "" {
				logger.Infof("Dispatching one-shot query mode with model=%s, stream=%v", model, stream)
				dispatchQuery(client)
			} else {
				logger.Infof("Entering interactive mode with model=%s, stream=%v", model, stream)
				interactiveMode(client)
			}
		},
	}

	rootCmd.Flags().StringVarP(&query, "query", "q", "", "Prompt to send to the model")
	rootCmd.Flags().StringVarP(&model, "model", "m", shared.ChatModelGPT4_1106Preview, "Model name")
	rootCmd.Flags().Float64VarP(&temperature, "temperature", "t", 0.7, "Sampling temperature")
	rootCmd.Flags().BoolVar(&markdown, "markdown", false, "Render output in Markdown")
	rootCmd.Flags().BoolVar(&stream, "stream", true, "Enable streaming response")
	rootCmd.Flags().StringVar(&baseURL, "base-url", "", "Custom OpenAI-compatible base URL")

	if err := rootCmd.Execute(); err != nil {
		logger.Fatalf("Command execution failed: %v", err)
	}
}

func dispatchQuery(client openai.Client) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(query),
	}
	sendRequest(client, messages)
}

func interactiveMode(client openai.Client) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("Interactive LLM Chat"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("36")).Render("You: "))
		scanner.Scan()
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" || line == "quit" {
			logger.Info("User exited interactive mode")
			break
		}
		logger.Debugf("User input: %s", line)
		messages := []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(line),
		}
		sendRequest(client, messages)
	}
}

func sendRequest(client openai.Client, messages []openai.ChatCompletionMessageParamUnion) {
	logger.Debugf("Sending request with %d message(s)...", len(messages))
	start := time.Now()

	// For this version of the library, we can't do streaming
	// Just note in the logs if stream was requested but unavailable
	if stream {
		logger.Warn("Streaming was requested but is not implemented in this version; using non-streaming API")
	}

	sendNonStreamRequest(client, messages, start)
}

func sendNonStreamRequest(client openai.Client, messages []openai.ChatCompletionMessageParamUnion, start time.Time) {
	logger.Debug("Using non-streaming mode")

	resp, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       model,
		Temperature: openai.Float(temperature),
	})
	if err != nil {
		logger.Fatalf("Chat completion failed: %v", err)
	}

	duration := time.Since(start)
	logger.Infof("Response received in %s", duration)

	if len(resp.Choices) > 0 && resp.Choices[0].Message.Content != "" {
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")).Render("Assistant:"))
		fmt.Println(resp.Choices[0].Message.Content)
		logger.Debugf("Response length: %d characters", len(resp.Choices[0].Message.Content))
	} else {
		logger.Warn("Received empty response from API")
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("31")).Render("Empty response received"))
	}
}
