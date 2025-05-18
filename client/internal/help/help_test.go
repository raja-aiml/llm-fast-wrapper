package help

import (
   "fmt"
   "os"
   "testing"

   "github.com/charmbracelet/glamour"
   "github.com/stretchr/testify/assert"
   "github.com/stretchr/testify/require"
)

func TestLoadUsageMarkdownInvalidPath(t *testing.T) {
   out := LoadUsageMarkdown("nonexistent.md")
   assert.Contains(t, out, "❌ Failed to load help markdown:")
}

func TestLoadUsageMarkdownValidPath(t *testing.T) {
   // usage.md is present in this directory
   out := LoadUsageMarkdown("usage.md")
   assert.NotEmpty(t, out)
   assert.NotContains(t, out, "❌ Failed to load help markdown:")
}
// badRenderer is a dummy UsageRenderer that always errors on Render
type badRenderer struct{}

func (b *badRenderer) Render(s string) (string, error) {
   return "", fmt.Errorf("render fail")
}
// Test when creating renderer fails: should return raw content
func TestLoadUsageMarkdownRendererInitError(t *testing.T) {
   orig := newHelpRenderer
   defer func() { newHelpRenderer = orig }()
   newHelpRenderer = func(opts ...glamour.TermRendererOption) (UsageRenderer, error) {
       return nil, fmt.Errorf("init fail")
   }
   data, err := os.ReadFile("usage.md")
   require.NoError(t, err)
   out := LoadUsageMarkdown("usage.md")
   assert.Equal(t, string(data), out)
}
// Test when rendering fails: should return raw content
func TestLoadUsageMarkdownRendererRenderError(t *testing.T) {
   orig := newHelpRenderer
   defer func() { newHelpRenderer = orig }()
   // override to return dummy renderer that errors on Render
   newHelpRenderer = func(opts ...glamour.TermRendererOption) (UsageRenderer, error) {
       return &badRenderer{}, nil
   }
   data, err := os.ReadFile("usage.md")
   require.NoError(t, err)
   out := LoadUsageMarkdown("usage.md")
   assert.Equal(t, string(data), out)
}