package audit

import (
   "bytes"
   "encoding/json"
   "fmt"
   "io"
   "os"
   "strings"
   "testing"
   "time"

   "github.com/stretchr/testify/require"
   "github.com/raja.aiml/llm-fast-wrapper/internal/config"
)

func TestLogAuditWritesJSONEntry(t *testing.T) {
   tmpFile, err := os.CreateTemp("", "audit-*.log")
   require.NoError(t, err)
   defer os.Remove(tmpFile.Name())

   cfg := &config.CLIConfig{
       Model:        "test-model",
       LogFile:      tmpFile.Name(),
       AuditEnabled: true,
   }

   prompt := "Hello"
   response := "World"
   LogAudit(prompt, response, cfg)

   data, err := os.ReadFile(tmpFile.Name())
   require.NoError(t, err)

   lines := strings.Split(strings.TrimSpace(string(data)), "\n")
   require.Len(t, lines, 1)

   var entry map[string]interface{}
   err = json.Unmarshal([]byte(lines[0]), &entry)
   require.NoError(t, err)

   require.Equal(t, prompt, entry["prompt"])
   require.Equal(t, response, entry["response"])
   require.Equal(t, cfg.Model, entry["model"])

   ts, ok := entry["time"].(string)
   require.True(t, ok)
   _, err = time.Parse(time.RFC3339, ts)
   require.NoError(t, err)
}
// captureStderr redirects stderr and returns captured output.
func captureStderr(f func()) string {
   old := os.Stderr
   r, w, _ := os.Pipe()
   os.Stderr = w
   f()
   w.Close()
   var buf bytes.Buffer
   io.Copy(&buf, r)
   os.Stderr = old
   return buf.String()
}

func TestLogAuditMarshalError(t *testing.T) {
   orig := jsonMarshal
   defer func() { jsonMarshal = orig }()
   jsonMarshal = func(v any) ([]byte, error) {
       return nil, fmt.Errorf("marshal fail")
   }
   cfg := &config.CLIConfig{Model: "test", LogFile: ""}
   out := captureStderr(func() {
       LogAudit("p", "r", cfg)
   })
   require.Contains(t, out, "Failed to marshal audit entry")
}

func TestLogAuditOpenFileError(t *testing.T) {
   orig := openFile
   defer func() { openFile = orig }()
   openFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
       return nil, fmt.Errorf("open fail")
   }
   cfg := &config.CLIConfig{Model: "test", LogFile: "dummy.log"}
   out := captureStderr(func() {
       LogAudit("p", "r", cfg)
   })
   require.Contains(t, out, "Failed to open audit log file")
}