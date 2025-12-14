package logger

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

// log is the unified internal logging function that handles both context and non-context calls
// captureSource parameter controls whether to capture stack trace (expensive operation)
func log(ctx context.Context, logger *slog.Logger, level slog.Level, msg string, captureSource bool, args ...any) {
	if !logger.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	if captureSource {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:])
		pc = pcs[0]
	}

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	_ = logger.Handler().Handle(ctx, r)
}

// LogWithSource logs a message with source information (no context)
func LogWithSource(logger *slog.Logger, level slog.Level, msg string, args ...any) {
	log(context.Background(), logger, level, msg, true, args...)
}

// LogWithSourceCtx logs a message with source information and context
func LogWithSourceCtx(ctx context.Context, logger *slog.Logger, level slog.Level, msg string, args ...any) {
	log(ctx, logger, level, msg, true, args...)
}

// LogNoSource logs a message without source information (faster for high-frequency logging)
func LogNoSource(logger *slog.Logger, level slog.Level, msg string, args ...any) {
	log(context.Background(), logger, level, msg, false, args...)
}

// LogNoSourceCtx logs a message without source information and with context
func LogNoSourceCtx(ctx context.Context, logger *slog.Logger, level slog.Level, msg string, args ...any) {
	log(ctx, logger, level, msg, false, args...)
}
