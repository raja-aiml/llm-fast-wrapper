package main

import (
   "bytes"
   "fmt"
   "io"
   "net/http"
   "net/http/httptest"
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
// Test main with --help flag prints usage without error
func TestMainHelp(t *testing.T) {
   cwd, err := os.Getwd()
   require.NoError(t, err)
   defer os.Chdir(cwd)
   // change to project root so logging can write to logs/llm-client.log
   require.NoError(t, os.Chdir(".."))
   oldArgs := os.Args
   defer func() { os.Args = oldArgs }()
   os.Args = []string{"llm-client", "--help"}
   out := captureOutput(func() {
       main()
   })
   assert.Contains(t, out, "Usage:")
}
// Test main with "help" subcommand prints full usage markdown
func TestMainHelpSubcommand(t *testing.T) {
   cwd, err := os.Getwd()
   require.NoError(t, err)
   defer os.Chdir(cwd)
   require.NoError(t, os.Chdir(".."))
   oldArgs := os.Args
   defer func() { os.Args = oldArgs }()
   os.Args = []string{"llm-client", "help"}
   out := captureOutput(func() {
       main()
   })
   // should contain some markdown header from usage file
   require.Contains(t, out, "#")
}
// Test main with --query invokes RunQuery and prints stub output
func TestMainRunQuery(t *testing.T) {
   cwd, err := os.Getwd()
   require.NoError(t, err)
   defer os.Chdir(cwd)
   require.NoError(t, os.Chdir(".."))
   // start test HTTP server for sync response
   ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       w.Header().Set("Content-Type", "application/json")
       fmt.Fprint(w, `{"id":"1","choices":[{"message":{"role":"assistant","content":"hello"},"finish_reason":"stop","index":0}]}`)
   }))
   defer ts.Close()
   // set env var
   oldEnv := os.Getenv("OPENAI_API_KEY")
   defer os.Setenv("OPENAI_API_KEY", oldEnv)
   os.Setenv("OPENAI_API_KEY", "testkey")
   // set args to include query and base-url
   oldArgs := os.Args
   defer func() { os.Args = oldArgs }()
   os.Args = []string{"llm-client", "--query", "foo", "--base-url", ts.URL}
   out := captureOutput(func() {
       main()
   })
   assert.Contains(t, out, "hello")
}
// Test main interactive mode when no query flag is provided
func TestMainRunInteractive(t *testing.T) {
   cwd, err := os.Getwd()
   require.NoError(t, err)
   defer os.Chdir(cwd)
   require.NoError(t, os.Chdir(".."))
   // set env var
   oldEnv := os.Getenv("OPENAI_API_KEY")
   defer os.Setenv("OPENAI_API_KEY", oldEnv)
   os.Setenv("OPENAI_API_KEY", "testkey")
   // simulate EOF on stdin for interactive mode
   r, w, _ := os.Pipe()
   w.Close()
   oldStdin := os.Stdin
   defer func() { os.Stdin = oldStdin }()
   os.Stdin = r
   oldArgs := os.Args
   defer func() { os.Args = oldArgs }()
   os.Args = []string{"llm-client"}
   out := captureOutput(func() {
       main()
   })
   // should print interactive header and prompt
   assert.Contains(t, out, "Interactive LLM Chat")
   assert.Contains(t, out, "You:")
}