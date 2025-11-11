package logbundle

import (
	"context"
	"log/slog"

	"github.com/aeternitas-infinita/logbundle-go/internal/logger"
)

func DebugCtx(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, Log, slog.LevelDebug, msg, args...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, Log, slog.LevelInfo, msg, args...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, Log, slog.LevelWarn, msg, args...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	logger.LogWithSourceCtx(ctx, Log, slog.LevelError, msg, args...)
}
