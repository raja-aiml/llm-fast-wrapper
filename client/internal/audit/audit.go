package audit

import (
   "encoding/json"
   "fmt"
   "os"
   "strings"
   "time"

   "github.com/charmbracelet/glamour"
   "github.com/charmbracelet/lipgloss"
   "github.com/raja.aiml/llm-fast-wrapper/internal/config"
   "gopkg.in/yaml.v3"
)

func PrintResponse(content string, cfg *config.CLIConfig) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")).Render("Assistant:"))

	switch cfg.Output {
   case "markdown":
       // Render markdown to terminal
       if strings.TrimSpace(content) == "" {
           fmt.Println("No content to render.")
       } else {
           r, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
           if err != nil {
               fmt.Println("Failed to render markdown:", err)
               fmt.Println(content)
           } else {
               out, err := r.Render(content)
               if err != nil {
                   fmt.Println("Markdown rendering error:", err)
                   fmt.Println(content)
               } else {
                   fmt.Println(out)
               }
           }
       }
	case "json":
		out, _ := json.MarshalIndent(map[string]string{
			"response": content,
		}, "", "  ")
		fmt.Println(string(out))
	case "yaml":
		out, _ := yaml.Marshal(map[string]string{
			"response": content,
		})
		fmt.Println(string(out))
	default: // "text"
		fmt.Println(content)
	}

	logAudit(cfg.Query, content, cfg)
}

func logAudit(prompt, response string, cfg *config.CLIConfig) {
	entry := map[string]any{
		"time":     time.Now().Format(time.RFC3339),
		"model":    cfg.Model,
		"prompt":   prompt,
		"response": response,
	}
	data, _ := json.Marshal(entry)

	f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(data)
	f.Write([]byte("\n"))
}
