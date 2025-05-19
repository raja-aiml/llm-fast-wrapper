package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger configures a combined console + optional file logger.
// If no paths are provided, logs to logs/llm-client.log and stdout.
func InitLogger(outPaths ...string) *zap.SugaredLogger {
	var cores []zapcore.Core

	// Console output
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel)
	cores = append(cores, consoleCore)

	// Default to log file if none given
	if len(outPaths) == 0 {
		outPaths = []string{"logs/llm-client.log"}
	}

	// File outputs (skip stdout/stderr)
	for _, path := range outPaths {
		if path == "stdout" || path == "stderr" {
			continue
		}

		// Ensure log directory exists
		if dir := filepath.Dir(path); dir != "" && dir != "." {
			_ = os.MkdirAll(dir, 0o755)
		}

		// Set up JSON file encoder
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.TimeKey = "time"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		fileEncoder := zapcore.NewJSONEncoder(encoderCfg)

		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to open log file %s: %v\n", path, err)
			continue
		}
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(file), zapcore.InfoLevel)
		cores = append(cores, fileCore)
	}

	// Combine all outputs
	logger := zap.New(zapcore.NewTee(cores...))
	return logger.Sugar()
}
