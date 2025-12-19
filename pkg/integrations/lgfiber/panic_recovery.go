package lgfiber

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/aeternitas-infinita/logbundle-go/pkg/config"
	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
	"github.com/aeternitas-infinita/logbundle-go/pkg/handler"
	"github.com/getsentry/sentry-go"
)

// RecoverGoroutinePanic recovers from panics in goroutines and logs them with full context
// This function should be used as: defer RecoverGoroutinePanic(ctx, "goroutineName")
// For best results with Sentry, pass the Fiber hub: defer RecoverGoroutinePanic(ctx, "goroutineName", sentryHub)
// It captures panic details, logs them, and sends to Sentry if enabled
func RecoverGoroutinePanic(ctx context.Context, goroutineName string) {
	if r := recover(); r != nil {
		// Get hub from context, fallback to current
		hub := sentry.GetHubFromContext(ctx)
		if hub == nil {
			hub = sentry.CurrentHub()
		}

		info := recoverPanic(ctx, r, hub, func(scope *sentry.Scope, info *panicInfo) {
			scope.SetLevel(sentry.LevelFatal)
			scope.SetTag("error_source", "goroutine_panic_recovery")
			scope.SetTag("goroutine_name", goroutineName)
			scope.SetTag("handled", "false")

			scope.SetContext("goroutine_details", map[string]any{
				"goroutine_name": goroutineName,
			})

			scope.SetFingerprint([]string{
				"goroutine_panic",
				goroutineName,
				fmt.Sprintf("%v", r),
				info.errorLoc,
			})

			hub.AddBreadcrumb(&sentry.Breadcrumb{
				Type:     "error",
				Category: "goroutine_panic",
				Message:  fmt.Sprintf("Panic in goroutine '%s': %v", goroutineName, r),
				Level:    sentry.LevelFatal,
				Data: map[string]any{
					"recovered_value": fmt.Sprintf("%v", r),
					"goroutine_name":  goroutineName,
					"location":        info.errorLoc,
				},
			}, nil)
		})

		// Use middleware logger if configured, otherwise fall back to internal logger
		log := config.GetMiddlewareLogger()
		if log == nil {
			log = handler.GetInternalLogger()
		}

		logFields := append([]any{
			slog.String("goroutine_name", goroutineName),
		}, info.logFields()...)

		log.ErrorContext(ctx, "Unhandled panic in goroutine", logFields...)
	}
}

// recoverPanic handles panic recovery logic with Sentry reporting
func recoverPanic(ctx context.Context, r any, hub *sentry.Hub, enrichScope func(*sentry.Scope, *panicInfo)) *panicInfo {
	stackTrace := string(debug.Stack())
	errorLoc, file, line := extractErrorLocationWithDetails(stackTrace)

	info := &panicInfo{
		recoveredValue: r,
		stackTrace:     stackTrace,
		errorLoc:       errorLoc,
		file:           file,
		line:           line,
	}

	var sentryEventID *sentry.EventID

	if config.IsSentryEnabled() && hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("panic_recovered", "true")
			scope.SetContext("panic_details", map[string]any{
				"recovered_value": fmt.Sprintf("%v", r),
				"stack_trace":     core.TruncateString(stackTrace, 5000),
				"error_location":  errorLoc,
			})

			if file != "" && line > 0 {
				scope.SetTag("panic_file", file)
				scope.SetTag("panic_line", fmt.Sprintf("%d", line))
				scope.SetContext("source", map[string]any{
					"file": file,
					"line": line,
				})
			}

			enrichScope(scope, info)
			sentryEventID = hub.CaptureException(fmt.Errorf("panic: %v", r))
		})
	}

	info.sentryEventID = sentryEventID
	return info
}

type panicInfo struct {
	recoveredValue any
	stackTrace     string
	errorLoc       string
	file           string
	line           int
	sentryEventID  *sentry.EventID
}

func (pi *panicInfo) logFields() []any {
	fields := []any{
		slog.Any("panic_value", pi.recoveredValue),
		slog.String("error_location", pi.errorLoc),
		slog.String("stack_trace", core.TruncateString(pi.stackTrace, 5000)),
	}

	if pi.sentryEventID != nil {
		fields = append(fields, slog.String("sentry_event_id", string(*pi.sentryEventID)))
	}

	if pi.file != "" && pi.line > 0 {
		fields = append(fields, slog.Any("source", slog.Source{
			File: pi.file,
			Line: pi.line,
		}))
	}

	return fields
}
