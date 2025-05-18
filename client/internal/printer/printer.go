package printer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// dependency injection for testing
var (
   newTermRenderer   = glamour.NewTermRenderer
   jsonMarshalIndent = json.MarshalIndent
   yamlMarshal       = yaml.Marshal
)

// Print renders model response based on format: markdown, json, yaml, or plain text.
func Print(content, format string) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")).Render("Assistant:"))

	switch strings.ToLower(format) {
	case "markdown":
		renderMarkdown(content)
	case "json":
		printJSON(content)
	case "yaml":
		printYAML(content)
	default:
		fmt.Println(content)
	}
}

func renderMarkdown(content string) {
	if strings.TrimSpace(content) == "" {
		fmt.Println("No content to render.")
		return
	}
	r, err := glamour.NewTermRenderer(glamour.WithAutoStyle())
	if err != nil {
		fmt.Println("Failed to render markdown:", err)
		fmt.Println(content)
		return
	}
	out, err := r.Render(content)
	if err != nil {
		fmt.Println("Markdown rendering error:", err)
		fmt.Println(content)
		return
	}
	fmt.Println(out)
}

func printJSON(content string) {
	out, err := json.MarshalIndent(map[string]string{"response": content}, "", "  ")
	if err != nil {
		fmt.Println("JSON encoding error:", err)
		fmt.Println(content)
		return
	}
	fmt.Println(string(out))
}

func printYAML(content string) {
	out, err := yaml.Marshal(map[string]string{"response": content})
	if err != nil {
		fmt.Println("YAML encoding error:", err)
		fmt.Println(content)
		return
	}
	fmt.Println(string(out))
}
