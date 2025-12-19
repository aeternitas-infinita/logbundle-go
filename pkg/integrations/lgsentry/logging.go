package lgsentry

import (
	"context"
	"log/slog"

	"github.com/aeternitas-infinita/logbundle-go/internal/logger"
	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
	"github.com/getsentry/sentry-go"
)

// Debug logs a debug message to slog and captures it in Sentry
func Debug(ctx context.Context, log *slog.Logger, msg string, extraData ...any) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	logger.LogWithSourceCtx(ctx, log, slog.LevelDebug, msg, extraData...)
	CaptureEvent(ctx, sentry.LevelDebug, msg, nil, extraData...)
}

// Info logs an info message to slog and captures it in Sentry
func Info(ctx context.Context, log *slog.Logger, msg string, extraData ...any) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	logger.LogWithSourceCtx(ctx, log, slog.LevelInfo, msg, extraData...)
	CaptureEvent(ctx, sentry.LevelInfo, msg, nil, extraData...)
}

// Warn logs a warning message to slog and captures it in Sentry
func Warn(ctx context.Context, log *slog.Logger, msg string, err error, extraData ...any) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if err != nil {
		allArgs := make([]any, 0, len(extraData)+1)
		allArgs = append(allArgs, core.ErrAttr(err))
		allArgs = append(allArgs, extraData...)
		logger.LogWithSourceCtx(ctx, log, slog.LevelWarn, msg, allArgs...)
	} else {
		logger.LogWithSourceCtx(ctx, log, slog.LevelWarn, msg, extraData...)
	}

	CaptureEvent(ctx, sentry.LevelWarning, msg, err, extraData...)
}

// Error logs an error message to slog and captures it in Sentry
func Error(ctx context.Context, log *slog.Logger, msg string, err error, extraData ...any) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	if err != nil {
		allArgs := make([]any, 0, len(extraData)+1)
		allArgs = append(allArgs, core.ErrAttr(err))
		allArgs = append(allArgs, extraData...)
		logger.LogWithSourceCtx(ctx, log, slog.LevelError, msg, allArgs...)
	} else {
		logger.LogWithSourceCtx(ctx, log, slog.LevelError, msg, extraData...)
	}

	CaptureEvent(ctx, sentry.LevelError, msg, err, extraData...)
}
