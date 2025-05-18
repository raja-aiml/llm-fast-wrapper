package chat

import (
   "bytes"
   "io"
   "os"
   "strings"
   "testing"

   "github.com/stretchr/testify/assert"
)

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

func TestRenderMarkdownEmpty(t *testing.T) {
   out := captureOutput(func() {
       RenderMarkdown("")
   })
   assert.Contains(t, out, "No content to render.")
}

func TestRenderMarkdownNonEmpty(t *testing.T) {
   markdown := "**bold**"
   out := captureOutput(func() {
       RenderMarkdown(markdown)
   })
   assert.NotContains(t, out, "No content to render.")
   assert.NotEmpty(t, strings.TrimSpace(out))
}