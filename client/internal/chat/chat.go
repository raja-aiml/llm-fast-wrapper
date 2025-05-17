package chat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/openai/openai-go"
	"go.uber.org/zap"

	"github.com/raja.aiml/llm-fast-wrapper/internal/config"
)

func RunQuery(client *openai.Client, cfg *config.CLIConfig, logger *zap.SugaredLogger) {
	logger.Infof("Running query with stream=%v", cfg.Stream)
	if cfg.Stream {
		runStreaming(client, cfg, logger)
	} else {
		runSync(client, cfg, logger)
	}
}

func RunInteractive(client *openai.Client, cfg *config.CLIConfig, logger *zap.SugaredLogger) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("Interactive LLM Chat"))
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("36")).Render("You: "))
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "exit" || input == "quit" {
			logger.Info("Exiting interactive mode")
			break
		}
		cfg.Query = input
		RunQuery(client, cfg, logger)
	}
}

func runSync(client *openai.Client, cfg *config.CLIConfig, logger *zap.SugaredLogger) {
   ctx := context.Background()
   // interactive spinner while thinking
   style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("34"))
   done := make(chan struct{})
   go func() {
       spinner := []rune{'|', '/', '-', '\\'}
       i := 0
       for {
           select {
           case <-done:
               return
           default:
               fmt.Printf("\r%s %c", style.Render("Thinking..."), spinner[i%len(spinner)])
               i++
               time.Sleep(200 * time.Millisecond)
           }
       }
   }()
	req := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(cfg.Model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(cfg.Query),
		},
		Temperature: openai.Float(cfg.Temperature),
	}
   resp, err := client.Chat.Completions.New(ctx, req)
   // stop spinner
   close(done)
   // clear spinner line
   fmt.Printf("\r")
	if err != nil {
		logger.Fatalf("OpenAI call failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		logger.Warn("Received empty response from OpenAI")
		return
	}

	content := resp.Choices[0].Message.Content
	printResponse(content, cfg.Markdown)
	logger.Debugf("Response length: %d characters", len(content))
}

func runStreaming(client *openai.Client, cfg *config.CLIConfig, logger *zap.SugaredLogger) {
	ctx := context.Background()

	params := openai.ChatCompletionNewParams{
		Model: openai.ChatModel(cfg.Model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(cfg.Query),
		},
		Temperature: openai.Float(cfg.Temperature),
	}
	stream := client.Chat.Completions.NewStreaming(ctx, params)
	defer stream.Close()

		// Print assistant label once, then stream content
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")).Render("Assistant:"))
		var fullText string
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) > 0 {
				text := chunk.Choices[0].Delta.Content
				fullText += text
				fmt.Print(text)
			}
		}
		if err := stream.Err(); err != nil {
			logger.Fatalf("Streaming error: %v", err)
		}
		// finish line after streaming
		fmt.Println()
		logger.Debugf("Streaming response length: %d characters", len(fullText))
}

func printResponse(content string, markdown bool) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")).Render("Assistant:"))
	if markdown {
		renderMarkdown(content)
	} else {
		fmt.Println(content)
	}
}

func renderMarkdown(text string) {
	r, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		fmt.Println("Failed to render markdown:", err)
		fmt.Println(text)
		return
	}
	out, err := r.Render(text)
	if err != nil {
		fmt.Println("Markdown rendering error:", err)
		fmt.Println(text)
		return
	}
	fmt.Println(out)
}
