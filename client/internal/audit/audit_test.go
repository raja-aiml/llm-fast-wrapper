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