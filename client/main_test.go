package main

import (
   "bytes"
   "io"
   "os"
   "testing"

   "github.com/stretchr/testify/assert"
   "github.com/stretchr/testify/require"
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
   // Change working directory so usage.md path resolves correctly
   cwd, err := os.Getwd()
   require.NoError(t, err)
   defer os.Chdir(cwd)
   // move to project root (one above client directory)
   require.NoError(t, os.Chdir(".."))
   cmd := helpCommand()
   out := captureOutput(func() {
       cmd.Run(cmd, []string{})
   })
   assert.NotEmpty(t, out)
   assert.NotContains(t, out, "❌ Failed")
}