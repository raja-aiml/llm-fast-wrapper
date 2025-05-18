package printer

import (
   "bytes"
   "fmt"
   "io"
   "os"
   "strings"
   "testing"

   "github.com/charmbracelet/glamour"
   "github.com/stretchr/testify/assert"
   "github.com/stretchr/testify/require"
)
// badRenderer is a dummy markdownRenderer that always errors on Render
type badRenderer struct{}

func (b *badRenderer) Render(s string) (string, error) {
   return "", fmt.Errorf("render fail")
}

// captureOutput redirects stdout and returns captured output as string.
func captureOutput(f func()) string {
   old := os.Stdout
   r, w, _ := os.Pipe()
   os.Stdout = w
   f()
   w.Close()
   var buf bytes.Buffer
   io.Copy(&buf, r)
   os.Stdout = old
   return buf.String()
}

func TestPrintDefault(t *testing.T) {
   out := captureOutput(func() {
       Print("hello", "")
   })
   assert.Contains(t, out, "Assistant:")
   assert.Contains(t, out, "hello")
}

func TestPrintJSON(t *testing.T) {
   out := captureOutput(func() {
       Print("hello", "json")
   })
   assert.Contains(t, out, `"response": "hello"`)
}

func TestPrintYAML(t *testing.T) {
   out := captureOutput(func() {
       Print("hello", "yaml")
   })
   assert.Contains(t, out, "response: hello")
}

func TestPrintMarkdownEmpty(t *testing.T) {
   out := captureOutput(func() {
       Print("", "markdown")
   })
   assert.Contains(t, out, "No content to render.")
}

func TestPrintCaseInsensitive(t *testing.T) {
   out := captureOutput(func() {
       Print("world", "JSON")
   })
   assert.Contains(t, out, `"response": "world"`)

   out2 := captureOutput(func() {
       Print("bar", "YAML")
   })
   assert.Contains(t, out2, "response: bar")
}
// Test non-empty markdown rendering branch
func TestPrintMarkdownNonEmpty(t *testing.T) {
   out := captureOutput(func() {
       Print("**bold**", "markdown")
   })
   // should render markdown without empty message
   assert.NotContains(t, out, "No content to render.")
   assert.NotEmpty(t, strings.TrimSpace(out))
}
// Test markdown renderer initialization error
func TestPrintMarkdownRendererInitError(t *testing.T) {
   orig := newTermRenderer
   defer func() { newTermRenderer = orig }()
   newTermRenderer = func(opts ...glamour.TermRendererOption) (markdownRenderer, error) {
       return nil, fmt.Errorf("init fail")
   }
   content := "**x**"
   out := captureOutput(func() {
       Print(content, "markdown")
   })
   require.Contains(t, out, "Failed to render markdown:")
   require.Contains(t, out, content)
}
// Test markdown renderer render error
func TestPrintMarkdownRendererRenderError(t *testing.T) {
   orig := newTermRenderer
   defer func() { newTermRenderer = orig }()
   newTermRenderer = func(opts ...glamour.TermRendererOption) (markdownRenderer, error) {
       return &badRenderer{}, nil
   }
   content := "**y**"
   out := captureOutput(func() {
       Print(content, "markdown")
   })
   require.Contains(t, out, "Markdown rendering error:")
   require.Contains(t, out, content)
}
// Test JSON encoding error branch
func TestPrintJSONError(t *testing.T) {
   orig := jsonMarshalIndent
   defer func() { jsonMarshalIndent = orig }()
   jsonMarshalIndent = func(v any, prefix, indent string) ([]byte, error) {
       return nil, fmt.Errorf("json error")
   }
   content := "foo"
   out := captureOutput(func() {
       Print(content, "json")
   })
   require.Contains(t, out, "JSON encoding error:")
   require.Contains(t, out, content)
}
// Test YAML encoding error branch
func TestPrintYAMLError(t *testing.T) {
   orig := yamlMarshal
   defer func() { yamlMarshal = orig }()
   yamlMarshal = func(v any) ([]byte, error) {
       return nil, fmt.Errorf("yaml error")
   }
   content := "bar"
   out := captureOutput(func() {
       Print(content, "yaml")
   })
   require.Contains(t, out, "YAML encoding error:")
   require.Contains(t, out, content)
}