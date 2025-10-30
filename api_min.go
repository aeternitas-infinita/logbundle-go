package logbundle

import (
	"context"
	"log/slog"

	"github.com/aeternitas-infinita/logbundle-go/internal/logger"
)

// DebugMin logs a debug message using the minimal logger (without source tracking)
func DebugMin(msg string, args ...any) {
	logger.LogWithSource(LogMin, slog.LevelDebug, msg, args...)
}

// InfoMin logs an informational message using the minimal logger (without source tracking)
func InfoMin(msg string, args ...any) {
	logger.LogWithSource(LogMin, slog.LevelInfo, msg, args...)
}

// WarnMin logs a warning message using the minimal logger (without source tracking)
func WarnMin(msg string, args ...any) {
	logger.LogWithSource(LogMin, slog.LevelWarn, msg, args...)
}

// ErrorMin logs an error message using the minimal logger (without source tracking)
func ErrorMin(msg string, args ...any) {
	logger.LogWithSource(LogMin, slog.LevelError, msg, args...)
}

// DebugCtxMin logs a debug message with context using the minimal logger
func DebugCtxMin(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, LogMin, slog.LevelDebug, msg, args...)
}

// InfoCtxMin logs an informational message with context using the minimal logger
func InfoCtxMin(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, LogMin, slog.LevelInfo, msg, args...)
}

// WarnCtxMin logs a warning message with context using the minimal logger
func WarnCtxMin(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, LogMin, slog.LevelWarn, msg, args...)
}

// ErrorCtxMin logs an error message with context using the minimal logger
func ErrorCtxMin(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, LogMin, slog.LevelError, msg, args...)
}
