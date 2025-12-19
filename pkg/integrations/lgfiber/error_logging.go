package lgfiber

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aeternitas-infinita/logbundle-go/pkg/config"
	"github.com/aeternitas-infinita/logbundle-go/pkg/handler"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
)

// logError logs an error with appropriate level and context
func logError(ctx context.Context, lgErr *lgerr.Error, sentryEventID *sentry.EventID, fiberCtx *fiber.Ctx) {
	// Use middleware logger if configured, otherwise fall back to internal logger
	log := config.GetMiddlewareLogger()
	if log == nil {
		log = handler.GetInternalLogger()
	}
	statusCode := lgErr.HTTPStatus()

	// Build log fields
	logFields := []any{
		slog.Int("status_code", statusCode),
		slog.String("error_type", string(lgErr.Type())),
		slog.String("error_message", lgErr.Message()),
	}

	// Add request info if available
	if fiberCtx != nil {
		logFields = append(logFields,
			slog.String("url", fiberCtx.OriginalURL()),
			slog.String("method", fiberCtx.Method()),
			slog.String("route", fiberCtx.Route().Path),
		)
	}

	// Add error context
	if errCtx := lgErr.Context(); len(errCtx) > 0 {
		logFields = append(logFields, slog.Any("error_context", errCtx))
	}

	// Add source location
	if lgErr.File() != "" && lgErr.Line() > 0 {
		logFields = append(logFields, slog.Any("source", slog.Source{
			File: lgErr.File(),
			Line: lgErr.Line(),
		}))
	}

	// Add Sentry event ID if captured
	if sentryEventID != nil {
		logFields = append(logFields, slog.String("sentry_event_id", string(*sentryEventID)))
	}

	// Add wrapped error
	if wrapped := lgErr.Wrapped(); wrapped != nil {
		logFields = append(logFields,
			slog.String("wrapped_error", wrapped.Error()),
			slog.String("wrapped_error_type", fmt.Sprintf("%T", wrapped)),
		)
	}

	// Add stack trace for server errors
	if statusCode >= 500 {
		if stackTrace := lgErr.FormatStackTrace(); stackTrace != "" {
			logFields = append(logFields, slog.String("stack_trace", stackTrace))
		}
	}

	// Log with appropriate level
	if statusCode >= 500 {
		log.ErrorContext(ctx, "Server error", logFields...)
	} else if statusCode >= 400 {
		log.WarnContext(ctx, "Client error", logFields...)
	} else {
		log.InfoContext(ctx, "Error handled", logFields...)
	}
}
