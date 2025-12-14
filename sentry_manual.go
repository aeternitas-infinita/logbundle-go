package logbundle

import (
	"context"
	"log/slog"

	"github.com/aeternitas-infinita/logbundle-go/internal/logger"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgsentry"
	"github.com/getsentry/sentry-go"
)

func SentryDebug(ctx context.Context, log *slog.Logger, msg string, extraData ...any) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	allArgs := make([]any, 0, len(extraData)+1)
	allArgs = append(allArgs, extraData...)
	logger.LogWithSourceCtx(ctx, log, slog.LevelDebug, msg, allArgs...)

	lgsentry.CaptureEvent(ctx, sentry.LevelDebug, msg, nil, extraData...)
}

func SentryInfo(ctx context.Context, log *slog.Logger, msg string, extraData ...any) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	allArgs := make([]any, 0, len(extraData)+1)
	allArgs = append(allArgs, extraData...)
	logger.LogWithSourceCtx(ctx, log, slog.LevelInfo, msg, allArgs...)

	lgsentry.CaptureEvent(ctx, sentry.LevelInfo, msg, nil, extraData...)
}

func SentryWarn(ctx context.Context, log *slog.Logger, msg string, err error, extraData ...any) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	allArgs := make([]any, 0, len(extraData)+1)
	if err != nil {
		allArgs = append(allArgs, ErrAttr(err))
	}
	allArgs = append(allArgs, extraData...)
	logger.LogWithSourceCtx(ctx, log, slog.LevelWarn, msg, allArgs...)

	lgsentry.CaptureEvent(ctx, sentry.LevelWarning, msg, err, extraData...)
}

func SentryError(ctx context.Context, log *slog.Logger, msg string, err error, extraData ...any) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	allArgs := make([]any, 0, len(extraData)+1)
	if err != nil {
		allArgs = append(allArgs, ErrAttr(err))
	}
	allArgs = append(allArgs, extraData...)
	logger.LogWithSourceCtx(ctx, log, slog.LevelError, msg, allArgs...)

	lgsentry.CaptureEvent(ctx, sentry.LevelError, msg, err, extraData...)
}
