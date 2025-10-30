package logbundle

import (
	"context"
	"log/slog"

	"github.com/aeternitas-infinita/logbundle-go/internal/logger"
)

// DebugCtx logs a debug message with context using the default logger
func DebugCtx(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, Log, slog.LevelDebug, msg, args...)
}

// InfoCtx logs an informational message with context using the default logger
func InfoCtx(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, Log, slog.LevelInfo, msg, args...)
}

// WarnCtx logs a warning message with context using the default logger
func WarnCtx(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, Log, slog.LevelWarn, msg, args...)
}

// ErrorCtx logs an error message with context using the default logger
func ErrorCtx(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, Log, slog.LevelError, msg, args...)
}
