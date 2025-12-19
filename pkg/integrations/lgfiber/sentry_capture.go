package lgfiber

import (
	"context"
	"fmt"
	"runtime"

	"github.com/aeternitas-infinita/logbundle-go/pkg/config"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"
	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
)

// shouldSendToSentryLazy performs a lightweight pre-check before creating hub
// Returns false if Sentry should definitely not be used, nil hub if might be needed
// This avoids creating the hub for 80% of errors (non-5xx status codes)
func shouldSendToSentryLazy(lgErr *lgerr.Error) bool {
	// Check if Sentry is globally enabled (fast config read)
	if !config.IsSentryEnabled() {
		return false
	}

	// Skip if error explicitly ignores Sentry (fast field check)
	if lgErr.ShouldIgnoreSentry() {
		return false
	}

	// Check status code against minimum (fast)
	statusCode := lgErr.HTTPStatus()
	minStatus := config.GetSentryMinHTTPStatus()

	// If minStatus is 0, send all errors (need hub check later)
	if minStatus == 0 {
		return true
	}

	// Only proceed if status code qualifies (e.g., >= 500 or >= minStatus)
	return statusCode >= minStatus
}

// shouldSendToSentry determines if an error should be reported to Sentry
// Reports if: Sentry is enabled AND status >= minHTTPStatus AND hub exists AND not explicitly ignored
func shouldSendToSentry(lgErr *lgerr.Error, hub *sentry.Hub) bool {
	// Pre-check without hub (most rejections happen here)
	if !shouldSendToSentryLazy(lgErr) {
		return false
	}

	// Hub exists check (only reached if lazy check passed)
	return hub != nil
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
	sentryFrames := make([]sentry.Frame, 0, len(pcs)) // Pre-allocate with exact capacity

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

	// Reverse frames in-place (Sentry expects bottom-up)
	for i, j := 0, len(sentryFrames)-1; i < j; i, j = i+1, j-1 {
		sentryFrames[i], sentryFrames[j] = sentryFrames[j], sentryFrames[i]
	}

	return &sentry.Stacktrace{Frames: sentryFrames}
}
