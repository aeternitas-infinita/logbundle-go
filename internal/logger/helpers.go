package logger

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

// LogWithSource logs with proper source location tracking
// skip=3 to bypass: Callers -> LogWithSource -> wrapper (Info/Debug/etc)
func LogWithSource(logger *slog.Logger, level slog.Level, msg string, args ...any) {
	// Early return optimization - avoid expensive runtime.Callers call
	if !logger.Enabled(context.Background(), level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip: Callers, LogWithSource, wrapper func (Info/Debug/etc)
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = logger.Handler().Handle(context.Background(), r)
}

// LogWithSourceCtx logs with context and proper source location
// skip=3 to bypass: Callers -> LogWithSourceCtx -> wrapper (InfoCtx/DebugCtx/etc)
func LogWithSourceCtx(ctx context.Context, logger *slog.Logger, level slog.Level, msg string, args ...any) {
	// Early return optimization - avoid expensive runtime.Callers call
	if !logger.Enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip: Callers, LogWithSourceCtx, wrapper func (InfoCtx/etc)
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = logger.Handler().Handle(ctx, r)
}
