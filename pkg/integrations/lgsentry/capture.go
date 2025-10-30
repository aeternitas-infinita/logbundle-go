package lgsentry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
)

// CaptureEventForSlog sends a slog record to Sentry based on filter configuration
// This is called automatically by the CustomHandler when Sentry is enabled
func CaptureEventForSlog(ctx context.Context, r slog.Record, args []slog.Attr) {
	// Early return if Sentry not initialized
	if !globalIntegration.initiated {
		return
	}

	config := globalIntegration.config

	// Check if this log level should be captured
	if !shouldCaptureLevel(r.Level, config.FilterLevels) {
		return
	}

	// Convert slog level to Sentry level
	sentryLevel := convertLogLevelToSentry(r.Level)

	// Extract and organize log data
	tags, extra, errorValue := extractSentryData(args)

	// Add trace ID for log correlation
	if traceID := core.GetLogTraceID(ctx); traceID != "" {
		tags[core.TraceIDKey] = traceID
	}

	// Add standard log metadata
	tags["log_level"] = r.Level.String()
	extra["timestamp"] = r.Time.Format(time.RFC3339)

	// Add source file/line information
	if sourceInfo := extractSourceInfo(r); sourceInfo != nil {
		tags["source"] = fmt.Sprintf("%s:%d", sourceInfo.File, sourceInfo.Line)
		extra["source_file"] = sourceInfo.File
		extra["source_line"] = sourceInfo.Line
	}

	// Try to get Fiber context and Sentry hub from context
	var hub *sentry.Hub
	var fiberCtx *fiber.Ctx

	// Extract fiber.Ctx if it exists in context (set by ContextEnrichmentMiddleware)
	if ctx != nil {
		if fc, ok := ctx.Value("fiber_ctx").(*fiber.Ctx); ok {
			fiberCtx = fc
			hub = sentryfiber.GetHubFromContext(fc)
		}
	}

	// Capture function that enriches scope and sends to Sentry
	captureFunc := func(scope *sentry.Scope) {
		scope.SetLevel(sentryLevel)

		// Set all tags for filtering/searching in Sentry
		for key, value := range tags {
			scope.SetTag(key, value)
		}

		// Set extra data for detailed investigation
		for key, value := range extra {
			scope.SetExtra(key, value)
		}

		// Set structured log context
		scope.SetContext("log_context", map[string]any{
			"message":   r.Message,
			"level":     r.Level.String(),
			"timestamp": r.Time.Format(time.RFC3339),
			"source":    tags["source"],
		})

		// Add request context if called from Fiber handler
		if fiberCtx != nil {
			scope.SetContext("request", map[string]any{
				"url":        fiberCtx.OriginalURL(),
				"method":     fiberCtx.Method(),
				"path":       fiberCtx.Path(),
				"route":      fiberCtx.Route().Path,
				"ip":         fiberCtx.IP(),
				"user_agent": fiberCtx.Get("User-Agent"),
			})

			// Add breadcrumb for this log entry
			hub.AddBreadcrumb(&sentry.Breadcrumb{
				Category:  "log",
				Message:   r.Message,
				Level:     sentryLevel,
				Timestamp: r.Time,
				Data:      extra,
			}, nil)
		}

		// Capture the event (exception if error present, otherwise message)
		if errorValue != nil {
			scope.SetTag("error_captured", "true")
			if hub != nil {
				hub.CaptureException(errorValue)
			} else {
				sentry.CaptureException(errorValue)
			}
		} else {
			if hub != nil {
				hub.CaptureMessage(r.Message)
			} else {
				sentry.CaptureMessage(r.Message)
			}
		}
	}

	// Use request-scoped hub if available, otherwise use global scope
	if hub != nil {
		hub.WithScope(captureFunc)
	} else {
		sentry.WithScope(captureFunc)
	}
}
