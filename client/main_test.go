package main

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

func TestHelpCommandMetadata(t *testing.T) {
   cmd := helpCommand()
   assert.Equal(t, "help", cmd.Use)
   assert.Equal(t, "Show full Markdown-based usage guide", cmd.Short)
}

func TestHelpCommandRun(t *testing.T) {
   cmd := helpCommand()
   out := captureOutput(func() {
       cmd.Run(cmd, []string{})
   })
   assert.NotEmpty(t, out)
   assert.NotContains(t, out, "❌ Failed")
}