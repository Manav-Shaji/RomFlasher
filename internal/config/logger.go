package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger creates a new Zap logger based on the provided configuration.
func InitLogger(cfg *AppConfig) (*zap.Logger, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 1. Ensure log directory exists
	logDir := cfg.Log.Dir
	if logDir == "" {
		logDir = "logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile := cfg.Log.Filename
	if logFile == "" {
		logFile = "app.log"
	}
	logPath := filepath.Join(logDir, logFile)

	// 2. Parse Log Level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Log.Level))); err != nil {
		level = zap.InfoLevel
	}

	// 3. Configure Encoder (JSON vs Console)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if strings.ToLower(cfg.Log.Format) == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 4. Setup Log Output (File)
	// Note: In a CLI/TUI app, logging to stdout directly will corrupt the UI.
	// Therefore, we strictly log to the file.
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", logPath, err)
	}

	writer := zapcore.AddSync(file)
	core := zapcore.NewCore(encoder, writer, level)

	logger := zap.New(core, zap.AddCaller())
	return logger, nil
}
