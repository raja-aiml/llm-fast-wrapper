package help

import (
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour"
)

// UsageRenderer defines the interface for rendering Markdown content.
type UsageRenderer interface {
   Render(string) (string, error)
}

// dependency injection hook for testing: override renderer creation.
var newHelpRenderer = func(opts ...glamour.TermRendererOption) (UsageRenderer, error) {
   return glamour.NewTermRenderer(opts...)
}
// LoadUsageMarkdown reads and renders the usage markdown file using glamour.
func LoadUsageMarkdown(path string) string {
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "❌ Failed to load help markdown: " + err.Error()
	}

	renderer, err := newHelpRenderer(
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
