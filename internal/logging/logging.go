package logging

import (
   "fmt"
   "os"
   "path/filepath"

   "go.uber.org/zap"
   "go.uber.org/zap/zapcore"
)

// InitLogger initializes a SugaredLogger writing to the given output paths.
// If no paths are provided, defaults to "logs/llm-client.log".
// Paths may be "stdout" or "stderr" or file paths; directories are created as needed.
func InitLogger(outPaths ...string) *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()
	// Determine output paths
	paths := outPaths
	if len(paths) == 0 {
		paths = []string{"logs/llm-client.log"}
	}
	// Ensure directories exist for file paths
	for _, p := range paths {
		if p != "stdout" && p != "stderr" {
			dir := filepath.Dir(p)
			if dir != "" && dir != "." {
				_ = os.MkdirAll(dir, 0o755)
			}
		}
	}
	cfg.OutputPaths = paths
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logr, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	return logr.Sugar()
}
