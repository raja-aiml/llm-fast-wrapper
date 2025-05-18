package printer

import (
   "bytes"
   "io"
   "os"
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