package help

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour"
)

// LoadUsageMarkdown reads and renders the usage markdown file using glamour.
func LoadUsageMarkdown(path string) string {
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "❌ Failed to load help markdown: " + err.Error()
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return string(content)
	}

	out, err := renderer.Render(string(content))
	if err != nil {
		return string(content)
	}
	return out
}
