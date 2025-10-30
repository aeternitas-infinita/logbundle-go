package logbundle

import (
	"log/slog"

	"github.com/aeternitas-infinita/logbundle-go/internal/logger"
)

// Info logs an informational message with the default logger
func Info(msg string, args ...any) {
	logger.LogWithSource(Log, slog.LevelInfo, msg, args...)
}

// Debug logs a debug message with the default logger
func Debug(msg string, args ...any) {
	logger.LogWithSource(Log, slog.LevelDebug, msg, args...)
}

// Warn logs a warning message with the default logger
func Warn(msg string, args ...any) {
	logger.LogWithSource(Log, slog.LevelWarn, msg, args...)
}

// Error logs an error message with the default logger
func Error(msg string, args ...any) {
	logger.LogWithSource(Log, slog.LevelError, msg, args...)
}
