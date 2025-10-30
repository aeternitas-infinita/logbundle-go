package logbundle

import (
	"context"
	"log/slog"
	"os"
	"runtime"
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
	h := handler.NewCustomHandler(os.Stdout, config.Level, config.AddSource, config.SentryEnabled)
	return slog.New(h)
}

func LogTraceIDToFHCtx(ctx *fasthttp.RequestCtx) {
	core.LogTraceIDToFHCtx(ctx)
}

func CtxWithLogTraceID(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return core.CtxWithLogTraceID(parent, timeout)
}

func GetLogTraceID(ctx any) string {
	return core.GetLogTraceID(ctx)
}

func ErrAttr(err error) slog.Attr {
	return core.ErrAttr(err)
}

func GetLvlFromStr(s string) slog.Level {
	return core.GetLvlFromStr(s)
}

func UpdateLogTraceIDKey(s string) {
	core.TraceIDKey = s
}

func GetBoolFromStr(s string) bool {
	return core.GetBoolFromStr(s)

}

// logWithSource logs with proper source location (skip = 3 to bypass this func and the wrapper)
func logWithSource(logger *slog.Logger, level slog.Level, msg string, args ...any) {
	if !logger.Enabled(context.Background(), level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip: Callers, logWithSource, wrapper func (Info/Debug/etc)
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = logger.Handler().Handle(context.Background(), r)
}

// logWithSourceCtx logs with context and proper source location
func logWithSourceCtx(ctx context.Context, logger *slog.Logger, level slog.Level, msg string, args ...any) {
	if !logger.Enabled(ctx, level) {
		return
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip: Callers, logWithSourceCtx, wrapper func (InfoCtx/etc)
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = logger.Handler().Handle(ctx, r)
}

func Info(msg string, args ...any) {
	logWithSource(Log, slog.LevelInfo, msg, args...)
}

func Debug(msg string, args ...any) {
	logWithSource(Log, slog.LevelDebug, msg, args...)
}

func Warn(msg string, args ...any) {
	logWithSource(Log, slog.LevelWarn, msg, args...)
}

func Error(msg string, args ...any) {
	logWithSource(Log, slog.LevelError, msg, args...)
}

func DebugCtx(ctx context.Context, msg string, args ...any) {
	logWithSourceCtx(ctx, Log, slog.LevelDebug, msg, args...)
}

func InfoCtx(ctx context.Context, msg string, args ...any) {
	logWithSourceCtx(ctx, Log, slog.LevelInfo, msg, args...)
}

func WarnCtx(ctx context.Context, msg string, args ...any) {
	logWithSourceCtx(ctx, Log, slog.LevelWarn, msg, args...)
}

func ErrorCtx(ctx context.Context, msg string, args ...any) {
	logWithSourceCtx(ctx, Log, slog.LevelError, msg, args...)
}

func DebugMin(msg string, args ...any) {
	logWithSource(LogMin, slog.LevelDebug, msg, args...)
}

func InfoMin(msg string, args ...any) {
	logWithSource(LogMin, slog.LevelInfo, msg, args...)
}

func WarnMin(msg string, args ...any) {
	logWithSource(LogMin, slog.LevelWarn, msg, args...)
}

func ErrorMin(msg string, args ...any) {
	logWithSource(LogMin, slog.LevelError, msg, args...)
}

func DebugCtxMin(ctx context.Context, msg string, args ...any) {
	logWithSourceCtx(ctx, LogMin, slog.LevelDebug, msg, args...)
}

func InfoCtxMin(ctx context.Context, msg string, args ...any) {
	logWithSourceCtx(ctx, LogMin, slog.LevelInfo, msg, args...)
}

func WarnCtxMin(ctx context.Context, msg string, args ...any) {
	logWithSourceCtx(ctx, LogMin, slog.LevelWarn, msg, args...)
}

func ErrorCtxMin(ctx context.Context, msg string, args ...any) {
	logWithSourceCtx(ctx, LogMin, slog.LevelError, msg, args...)
}
