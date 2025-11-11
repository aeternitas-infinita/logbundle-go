package logbundle

import (
	"log/slog"

	"github.com/aeternitas-infinita/logbundle-go/internal/logger"
)

func Info(msg string, args ...any) {
	logger.LogWithSource(Log, slog.LevelInfo, msg, args...)
}

func Debug(msg string, args ...any) {
	logger.LogWithSource(Log, slog.LevelDebug, msg, args...)
}

func Warn(msg string, args ...any) {
	logger.LogWithSource(Log, slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	logger.LogWithSource(Log, slog.LevelError, msg, args...)
}
