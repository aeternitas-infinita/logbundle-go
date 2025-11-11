package logbundle

import (
	"context"

	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgsentry"
	"github.com/getsentry/sentry-go"
)

func SentryDebug(ctx context.Context, msg string, err error, extraData ...any) {
	allArgs := make([]any, 0, len(extraData)+1)
	if err != nil {
		allArgs = append(allArgs, ErrAttr(err))
	}
	allArgs = append(allArgs, extraData...)
	DebugCtx(ctx, msg, allArgs...)

	lgsentry.CaptureEvent(ctx, sentry.LevelDebug, msg, err, extraData...)
}

func SentryInfo(ctx context.Context, msg string, err error, extraData ...any) {
	allArgs := make([]any, 0, len(extraData)+1)
	if err != nil {
		allArgs = append(allArgs, ErrAttr(err))
	}
	allArgs = append(allArgs, extraData...)
	InfoCtx(ctx, msg, allArgs...)

	lgsentry.CaptureEvent(ctx, sentry.LevelInfo, msg, err, extraData...)
}

func SentryWarn(ctx context.Context, msg string, err error, extraData ...any) {
	allArgs := make([]any, 0, len(extraData)+1)
	if err != nil {
		allArgs = append(allArgs, ErrAttr(err))
	}
	allArgs = append(allArgs, extraData...)
	WarnCtx(ctx, msg, allArgs...)

	lgsentry.CaptureEvent(ctx, sentry.LevelWarning, msg, err, extraData...)
}

func SentryError(ctx context.Context, msg string, err error, extraData ...any) {
	allArgs := make([]any, 0, len(extraData)+1)
	if err != nil {
		allArgs = append(allArgs, ErrAttr(err))
	}
	allArgs = append(allArgs, extraData...)
	ErrorCtx(ctx, msg, allArgs...)

	lgsentry.CaptureEvent(ctx, sentry.LevelError, msg, err, extraData...)
}
