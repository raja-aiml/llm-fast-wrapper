package chat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	openai "github.com/openai/openai-go"
	"go.uber.org/zap"

	"github.com/raja.aiml/llm-fast-wrapper/internal/config"
)

func RunQuery(client openai.Client, cfg *config.CLIConfig, logger *zap.SugaredLogger) {
	logger.Infof("Running one-shot query with stream=%v", cfg.Stream)
	send(client, cfg, logger, cfg.Query)
}

func RunInteractive(client openai.Client, cfg *config.CLIConfig, logger *zap.SugaredLogger) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("Interactive LLM Chat"))
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("36")).Render("You: "))
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" || line == "quit" {
			logger.Info("Exiting interactive mode")
			break
		}
		send(client, cfg, logger, line)
	}
}

func send(client openai.Client, cfg *config.CLIConfig, logger *zap.SugaredLogger, prompt string) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(prompt),
	}

	logger.Debugf("Sending request with prompt: %s", prompt)
	start := time.Now()

	if cfg.Stream {
		logger.Warn("Streaming not implemented; falling back to standard request")
	}

	resp, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       cfg.Model,
		Temperature: openai.Float(cfg.Temperature),
	})
	if err != nil {
		logger.Fatalf("Chat completion failed: %v", err)
	}

	duration := time.Since(start)
	logger.Infof("Response received in %s", duration)

	if len(resp.Choices) > 0 {
		content := resp.Choices[0].Message.Content
		if content != "" {
			fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")).Render("Assistant:"))
			fmt.Println(content)
			logger.Debugf("Response: %d characters", len(content))
		} else {
			logger.Warn("Received empty response")
		}
	} else {
		logger.Warn("No choices returned in response")
	}
}
