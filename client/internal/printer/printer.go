package printer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// markdownRenderer defines the interface for Markdown rendering backends.
type markdownRenderer interface {
   Render(string) (string, error)
}

// dependency injection hooks for testing: override rendering and serialization.
var (
   // newTermRenderer is the factory for our Markdown renderer
   newTermRenderer = func(opts ...glamour.TermRendererOption) (markdownRenderer, error) {
       return glamour.NewTermRenderer(opts...)
   }
   // jsonMarshalIndent serializes JSON responses
   jsonMarshalIndent = json.MarshalIndent
   // yamlMarshal serializes YAML responses
   yamlMarshal = yaml.Marshal
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
	r, err := newTermRenderer(glamour.WithAutoStyle())
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
   out, err := jsonMarshalIndent(map[string]string{"response": content}, "", "  ")
	if err != nil {
		fmt.Println("JSON encoding error:", err)
		fmt.Println(content)
		return
	}
	fmt.Println(string(out))
}

func printYAML(content string) {
   out, err := yamlMarshal(map[string]string{"response": content})
	if err != nil {
		fmt.Println("YAML encoding error:", err)
		fmt.Println(content)
		return
	}
	fmt.Println(string(out))
}
