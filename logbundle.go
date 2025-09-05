package logbundle

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
	"github.com/aeternitas-infinita/logbundle-go/pkg/handler"
)

var Log = slog.New(handler.NewCustomHandler(
	os.Stdout,
	core.GetLvlFromEnv("log_level"),
	true,
	false,
))

var LogMin = slog.New(handler.NewCustomHandler(
	os.Stdout,
	core.GetLvlFromEnv("log_level"),
	false,
	false,
))

func InitLog(cfg LoggerConfig) {
	Log = CreateLogger(cfg)
}

func InitLogMin(cfg LoggerConfig) {
	LogMin = CreateLogger(cfg)
}

type LoggerConfig struct {
	Level         slog.Level
	SentryEnabled bool
	AddSource     bool
}

func CreateLogger(config LoggerConfig) *slog.Logger {
	handler := handler.NewCustomHandler(os.Stdout, config.Level, config.AddSource, config.SentryEnabled)
	return slog.New(handler)
}

func TraceIDToFHCtx(ctx *fasthttp.RequestCtx) {
	core.TraceIDToFHCtx(ctx)
}

func CtxWithTraceID(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return core.CtxWithTraceID(parent, timeout)
}

func GetTraceID(ctx any) string {
	return core.GetTraceID(ctx)
}

func ErrAttr(err error) slog.Attr {
	return core.ErrAttr(err)
}

func GetLvlFromStr(s string) slog.Level {
	return core.GetLvlFromStr(s)
}

func UpdateTraceIDKey(s string) {
	core.TraceIDKey = s
}

func GetBoolFromStr(s string) bool {
	return core.GetBoolFromStr(s)

}

func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

func DebugCtx(ctx context.Context, msg string, args ...any) {
	Log.DebugContext(ctx, msg, args...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	Log.InfoContext(ctx, msg, args...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	Log.WarnContext(ctx, msg, args...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	Log.ErrorContext(ctx, msg, args...)
}

func DebugMin(msg string, args ...any) {
	LogMin.Debug(msg, args...)
}

func InfoMin(msg string, args ...any) {
	LogMin.Info(msg, args...)
}

func WarnMin(msg string, args ...any) {
	LogMin.Warn(msg, args...)
}

func ErrorMin(msg string, args ...any) {
	LogMin.Error(msg, args...)
}

func DebugCtxMin(ctx context.Context, msg string, args ...any) {
	LogMin.DebugContext(ctx, msg, args...)
}

func InfoCtxMin(ctx context.Context, msg string, args ...any) {
	LogMin.InfoContext(ctx, msg, args...)
}

func WarnCtxMin(ctx context.Context, msg string, args ...any) {
	LogMin.WarnContext(ctx, msg, args...)
}

func ErrorCtxMin(ctx context.Context, msg string, args ...any) {
	LogMin.ErrorContext(ctx, msg, args...)
}
