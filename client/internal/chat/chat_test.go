package chat

import (
   "bytes"
   "fmt"
   "io"
   "net/http"
   "net/http/httptest"
   "os"
   "strings"
   "testing"

   "github.com/stretchr/testify/assert"

   openai "github.com/openai/openai-go"
   "github.com/openai/openai-go/option"
   "github.com/raja.aiml/llm-fast-wrapper/internal/config"
   "go.uber.org/zap"
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
// Test RunInteractive exits on "exit" input without error
func TestRunInteractiveExit(t *testing.T) {
   // prepare input "exit\n" via pipe
   r, w, _ := os.Pipe()
   w.Write([]byte("exit\n"))
   w.Close()
   oldStdin := os.Stdin
   defer func() { os.Stdin = oldStdin }()
   os.Stdin = r
   cfg := &config.CLIConfig{}
   logger := zap.NewNop().Sugar()
   out := captureOutput(func() {
       RunInteractive(nil, cfg, logger)
   })
   assert.Contains(t, out, "Interactive LLM Chat")
   assert.Contains(t, out, "You:")
}
// Test RunQuery non-stream (sync) prints response
func TestRunQuerySync(t *testing.T) {
   // create a test server returning ChatCompletion JSON
   ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       w.Header().Set("Content-Type", "application/json")
       resp := `{"id":"1","choices":[{"message":{"role":"assistant","content":"hello"},"finish_reason":"stop","index":0}]}`
       fmt.Fprint(w, resp)
   }))
   defer ts.Close()
   // create client with baseURL and test HTTP client
   client := openai.NewClient(option.WithBaseURL(ts.URL+"/"), option.WithHTTPClient(ts.Client()))
   cfg := &config.CLIConfig{Model: "m", Query: "q", Temperature: 0.0, Stream: false, AuditEnabled: false, Output: "text"}
   logger := zap.NewNop().Sugar()
   out := captureOutput(func() {
       RunQuery(&client, cfg, logger)
   })
   assert.Contains(t, out, "hello")
}
// Test RunQuery streaming prints streamed chunks
func TestRunQueryStream(t *testing.T) {
   // create a test server returning streaming SSE
   ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
       w.Header().Set("Content-Type", "text/event-stream")
       // send two chunk events
       fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"h\"},\"index\":0}]}\n\n")
       fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"i\"},\"index\":0}]}\n\n")
   }))
   defer ts.Close()
   client := openai.NewClient(option.WithBaseURL(ts.URL+"/"), option.WithHTTPClient(ts.Client()))
   cfg := &config.CLIConfig{Model: "m", Query: "q", Temperature: 0.0, Stream: true, AuditEnabled: false, Output: "text"}
   logger := zap.NewNop().Sugar()
   out := captureOutput(func() {
       RunQuery(&client, cfg, logger)
   })
   // header and streamed content should appear
   assert.Contains(t, out, "Assistant:")
   assert.Contains(t, out, "hi")
}