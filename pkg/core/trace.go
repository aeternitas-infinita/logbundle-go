package core

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

// LogTraceIDToFHCtx generates and stores a new trace ID in fasthttp request context
func LogTraceIDToFHCtx(ctx *fasthttp.RequestCtx) {
	ctx.SetUserValue(TraceIDKey, uuid.New().String())
}

// CtxWithLogTraceID creates a new context with timeout and adds a trace ID
func CtxWithLogTraceID(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	return context.WithValue(ctx, TraceIDKey, uuid.New().String()), cancel
}

// GetLogTraceID retrieves trace ID from context (supports both fasthttp.RequestCtx and context.Context)
// Returns empty string if trace ID is not found or context is nil
func GetLogTraceID(ctx any) string {
	if ctx == nil {
		return ""
	}

	// Check fasthttp.RequestCtx first (more common in Fiber apps)
	if requestCtx, ok := ctx.(*fasthttp.RequestCtx); ok {
		if v := requestCtx.UserValue(TraceIDKey); v != nil {
			if traceID, ok := v.(string); ok {
				return traceID
			}
		}
		return ""
	}

	// Check standard context.Context
	if stdCtx, ok := ctx.(context.Context); ok {
		if v := stdCtx.Value(TraceIDKey); v != nil {
			if traceID, ok := v.(string); ok {
				return traceID
			}
		}
	}

	return ""
}
