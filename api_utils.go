package logbundle

import (
	"context"
	"log/slog"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
)

// ErrAttr creates a slog.Attr for an error with key "error"
func ErrAttr(err error) slog.Attr {
	return core.ErrAttr(err)
}

// GetLogTraceID retrieves trace ID from context (supports both fasthttp.RequestCtx and context.Context)
// Returns empty string if trace ID is not found or context is nil
func GetLogTraceID(ctx any) string {
	return core.GetLogTraceID(ctx)
}

// LogTraceIDToFHCtx generates and stores a new trace ID in fasthttp request context
func LogTraceIDToFHCtx(ctx *fasthttp.RequestCtx) {
	core.LogTraceIDToFHCtx(ctx)
}

// CtxWithLogTraceID creates a new context with timeout and adds a trace ID
func CtxWithLogTraceID(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return core.CtxWithLogTraceID(parent, timeout)
}

// GetLvlFromStr converts string to slog.Level
// Accepts: "debug", "info", "warn", "error" (case-sensitive)
// Returns slog.LevelWarn for invalid values
func GetLvlFromStr(s string) slog.Level {
	return core.GetLvlFromStr(s)
}

// UpdateLogTraceIDKey updates the global trace ID context key
// Use with caution - affects all subsequent trace ID operations
func UpdateLogTraceIDKey(s string) {
	core.TraceIDKey = s
}

// GetBoolFromStr converts string to boolean (case-insensitive)
// Returns true only if string is "true" (case-insensitive), false otherwise
func GetBoolFromStr(s string) bool {
	return core.GetBoolFromStr(s)
}
