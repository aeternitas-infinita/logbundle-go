package lgfiber

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/config"
	"github.com/aeternitas-infinita/logbundle-go/pkg/handler"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"
)

// shouldSendToSentry determines if an error should be reported to Sentry
// Reports if: Sentry is enabled AND status >= minHTTPStatus AND hub exists AND not explicitly ignored
func shouldSendToSentry(lgErr *lgerr.Error, hub *sentry.Hub) bool {
	// Check if Sentry is globally enabled
	if !config.IsSentryEnabled() {
		return false
	}

	if hub == nil || lgErr.ShouldIgnoreSentry() {
		return false
	}

	statusCode := lgErr.HTTPStatus()
	minStatus := config.GetSentryMinHTTPStatus()

	// If minStatus is 0, send all errors
	if minStatus == 0 {
		return true
	}

	return statusCode >= minStatus
}

// captureToSentry captures an lgerr.Error to Sentry with full context
func captureToSentry(ctx context.Context, hub *sentry.Hub, lgErr *lgerr.Error, source string, fiberCtx *fiber.Ctx) *sentry.EventID {
	if hub == nil {
		return nil
	}

	var eventID *sentry.EventID

	hub.WithScope(func(scope *sentry.Scope) {
		// Set basic tags
		scope.SetLevel(sentry.LevelError)
		scope.SetTag("error_source", source)
		scope.SetTag("error_type", string(lgErr.Type()))
		scope.SetTag("status_code", fmt.Sprintf("%d", lgErr.HTTPStatus()))

		// Add error context
		if errCtx := lgErr.Context(); len(errCtx) > 0 {
			scope.SetContext("error_context", errCtx)
		}

		// Add request context if available
		if fiberCtx != nil {
			scope.SetContext("request", map[string]any{
				"url":        fiberCtx.OriginalURL(),
				"method":     fiberCtx.Method(),
				"path":       fiberCtx.Path(),
				"route":      fiberCtx.Route().Path,
				"ip":         fiberCtx.IP(),
				"user_agent": fiberCtx.Get("User-Agent"),
			})

			if queries := fiberCtx.Queries(); len(queries) > 0 {
				queryParams := make(map[string]any)
				for k, v := range queries {
					queryParams[k] = v
				}
				scope.SetContext("query_params", queryParams)
			}
		}

		// Add source location if available
		if lgErr.File() != "" && lgErr.Line() > 0 {
			scope.SetTag("error_file", lgErr.File())
			scope.SetTag("error_line", fmt.Sprintf("%d", lgErr.Line()))
			scope.SetContext("source", map[string]any{
				"file": lgErr.File(),
				"line": lgErr.Line(),
			})
		}

		// Set fingerprint for grouping
		scope.SetFingerprint([]string{
			source,
			string(lgErr.Type()),
			lgErr.Message(),
		})

		// Build Sentry exception
		event := sentry.NewEvent()
		event.Level = sentry.LevelError
		event.Message = lgErr.Message()

		exception := sentry.Exception{
			Type:  fmt.Sprintf("lgerr.%s", lgErr.Type()),
			Value: lgErr.Error(),
			Mechanism: &sentry.Mechanism{
				Type:    "lgerr_handler",
				Handled: func() *bool { b := true; return &b }(),
			},
		}

		// Add stack trace if available
		if stackTrace := lgErr.StackTrace(); len(stackTrace) > 0 {
			exception.Stacktrace = buildStacktrace(stackTrace)
		}

		// Add wrapped error info
		if wrapped := lgErr.Wrapped(); wrapped != nil {
			if exception.Mechanism.Data == nil {
				exception.Mechanism.Data = make(map[string]any)
			}
			exception.Mechanism.Data["wrapped_error"] = wrapped.Error()
			exception.Mechanism.Data["wrapped_error_type"] = fmt.Sprintf("%T", wrapped)
		}

		event.Exception = []sentry.Exception{exception}
		eventID = hub.CaptureEvent(event)
	})

	return eventID
}

// buildStacktrace converts runtime stack trace to Sentry format
func buildStacktrace(pcs []uintptr) *sentry.Stacktrace {
	if len(pcs) == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pcs)
	var sentryFrames []sentry.Frame

	for {
		frame, more := frames.Next()
		sentryFrames = append(sentryFrames, sentry.Frame{
			Filename: frame.File,
			Function: frame.Function,
			Lineno:   frame.Line,
			AbsPath:  frame.File,
		})
		if !more {
			break
		}
	}

	// Reverse frames (Sentry expects bottom-up)
	for i, j := 0, len(sentryFrames)-1; i < j; i, j = i+1, j-1 {
		sentryFrames[i], sentryFrames[j] = sentryFrames[j], sentryFrames[i]
	}

	return &sentry.Stacktrace{Frames: sentryFrames}
}

// logError logs an error with appropriate level and context
func logError(ctx context.Context, lgErr *lgerr.Error, sentryEventID *sentry.EventID, fiberCtx *fiber.Ctx) {
	log := handler.GetInternalLogger()
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

// ErrorHandler is the main Fiber error handler
// Catches errors, logs them, and sends to Sentry if appropriate
func ErrorHandler(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	// Try to extract lgerr.Error
	var lgErr *lgerr.Error
	if !errors.As(err, &lgErr) {
		// Not an lgerr.Error - convert to lgerr.Internal for consistent handling
		code := fiber.StatusInternalServerError
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			code = fiberErr.Code
		}

		// Create lgerr.Error from generic error
		lgErr = lgerr.Internal(err.Error()).
			Wrap(err).
			WithHTTPStatus(code).
			WithTitle("Internal Server Error").
			WithContext("original_error_type", fmt.Sprintf("%T", err))

		// Continue with normal lgerr.Error handling flow
	}

	// Handle lgerr.Error
	hub := sentryfiber.GetHubFromContext(c)
	var sentryEventID *sentry.EventID

	// Send to Sentry if appropriate
	if shouldSendToSentry(lgErr, hub) {
		sentryEventID = captureToSentry(c.UserContext(), hub, lgErr, "error_handler", c)
	}

	// Log the error
	logError(c.UserContext(), lgErr, sentryEventID, c)

	// Return error response
	return c.Status(lgErr.HTTPStatus()).JSON(lgErr.ToErrorResponse())
}

// HandleError manually handles an lgerr.Error with logging and Sentry reporting
// Use this for explicit error handling in goroutines or background tasks
//
// Example usage in goroutine:
//
//	go func() {
//	    err := performBackgroundTask()
//	    if err != nil {
//	        lgErr := lgerr.Internal("background task failed").Wrap(err)
//	        lgfiber.HandleError(ctx, lgErr)
//	    }
//	}()
func HandleError(ctx context.Context, lgErr *lgerr.Error) *sentry.EventID {
	if lgErr == nil {
		return nil
	}

	hub := sentry.GetHubFromContext(ctx)
	var sentryEventID *sentry.EventID

	// Send to Sentry if appropriate
	if shouldSendToSentry(lgErr, hub) {
		sentryEventID = captureToSentry(ctx, hub, lgErr, "manual_handle", nil)
	}

	// Log the error
	logError(ctx, lgErr, sentryEventID, nil)

	return sentryEventID
}

// HandleErrorWithFiber manually handles an lgerr.Error with full Fiber context
// Use this for explicit error handling within Fiber handlers when you don't want to return the error
//
// Example usage:
//
//	func handler(c *fiber.Ctx) error {
//	    // Async operation
//	    go func() {
//	        if err := doSomething(); err != nil {
//	            lgErr := lgerr.Internal("operation failed").Wrap(err)
//	            lgfiber.HandleErrorWithFiber(c, lgErr)
//	        }
//	    }()
//
//	    return c.JSON(fiber.Map{"status": "processing"})
//	}
func HandleErrorWithFiber(c *fiber.Ctx, lgErr *lgerr.Error) *sentry.EventID {
	if lgErr == nil {
		return nil
	}

	hub := sentryfiber.GetHubFromContext(c)
	var sentryEventID *sentry.EventID

	// Send to Sentry if appropriate with full Fiber context
	if shouldSendToSentry(lgErr, hub) {
		sentryEventID = captureToSentry(c.UserContext(), hub, lgErr, "manual_fiber_handle", c)
	}

	// Log the error with Fiber context
	logError(c.UserContext(), lgErr, sentryEventID, c)

	return sentryEventID
}
