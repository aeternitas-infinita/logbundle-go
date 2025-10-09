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

	// Skip goroutine line and runtime panic lines
	// Looking for the first non-runtime, non-library line
	for i := 0; i < len(lines)-1; i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip goroutine info
		if strings.HasPrefix(line, "goroutine ") {
			continue
		}

		// Look at the next line which should contain file:line
		if i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])

			// Check if this line contains .go: (file path with line number)
			if strings.Contains(nextLine, ".go:") {
				// Skip runtime and library internals
				if strings.Contains(nextLine, "runtime/") ||
					strings.Contains(nextLine, "logbundle-go/") ||
					strings.Contains(line, "FiberRecoverMiddleware") ||
					strings.Contains(line, "RecoverWithContext") {
					continue
				}

				// Extract file path
				parts := strings.Fields(nextLine)
				if len(parts) > 0 {
					filePath := parts[0]
					return filePath
				}
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
