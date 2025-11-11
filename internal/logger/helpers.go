package logger

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

// log is the unified internal logging function that handles both context and non-context calls
func log(ctx context.Context, logger *slog.Logger, level slog.Level, msg string, args ...any) {
	if !logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = logger.Handler().Handle(ctx, r)
}

// LogWithSource logs a message with source information (no context)
func LogWithSource(logger *slog.Logger, level slog.Level, msg string, args ...any) {
	log(context.Background(), logger, level, msg, args...)
}

// LogWithSourceCtx logs a message with source information and context
func LogWithSourceCtx(ctx context.Context, logger *slog.Logger, level slog.Level, msg string, args ...any) {
	log(ctx, logger, level, msg, args...)
}
