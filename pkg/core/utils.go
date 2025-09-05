package core

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func TraceIDToFHCtx(ctx *fasthttp.RequestCtx) {
	ctx.SetUserValue(TraceIDKey, uuid.New().String())
}

func CtxWithTraceID(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	return context.WithValue(ctx, TraceIDKey, uuid.New().String()), cancel
}

func GetTraceID(ctx any) string {
	if ctx == nil {
		return ""
	}

	if requestCtx, ok := ctx.(*fasthttp.RequestCtx); ok {
		if v := requestCtx.UserValue(TraceIDKey); v != nil {
			return v.(string)
		}
		return ""
	}

	if stdCtx, ok := ctx.(context.Context); ok {
		if v := stdCtx.Value(TraceIDKey); v != nil {
			if traceID, ok := v.(string); ok {
				return traceID
			}
		}
	}

	return ""
}

func ExtractErrorLocation(stackTrace string) string {
	lines := strings.Split(stackTrace, "\n")

	for i := 0; i < len(lines)-1; i++ {
		if strings.Contains(lines[i], "im-in-fairy-tale-main") &&
			!strings.Contains(lines[i], "FiberRecoverMiddleware") {

			nextLine := ""
			if i+1 < len(lines) {
				nextLine = lines[i+1]
			}

			if strings.Contains(nextLine, ".go:") {
				filePath := strings.TrimSpace(nextLine)

				if idx := strings.LastIndex(filePath, "im-in-fairy-tale-main-backend"); idx != -1 {
					filePath = filePath[idx:]

					parts := strings.Split(filePath, " ")
					if len(parts) > 0 {
						cleanPath := parts[0]
						const prefix = "im-in-fairy-tale-main-backend/"
						cleanPath = strings.TrimPrefix(cleanPath, prefix)
						return cleanPath
					}
					return filePath
				}
				return filePath
			}
		}
	}

	return "unknown location"
}

func GetLvlFromEnv(key string) slog.Level {
	if value := os.Getenv(key); value != "" {
		return GetLvlFromStr(value)
	}
	return slog.LevelWarn
}

func GetLvlFromStr(s string) slog.Level {
	var level slog.Level

	switch s {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelWarn
	}

	return level
}

func GetLinePositionStringWithSkip(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("[%s:%d]", file, line)
}

func GetBoolFromStr(s string) bool {
	return strings.ToLower(s) == "true"
}
