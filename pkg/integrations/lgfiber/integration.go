package lgfiber

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
	"github.com/aeternitas-infinita/logbundle-go/pkg/handler"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/erri"
)

type userIDProvider interface {
	GetUserID() string
}

// captureErrorToSentry is a centralized helper for capturing errors to Sentry
// It reduces code duplication and ensures consistent error handling
func captureErrorToSentry(c *fiber.Ctx, err error, code int, source string) *sentry.EventID {
	hub := sentryfiber.GetHubFromContext(c)
	if hub == nil {
		return nil
	}

	var eventID *sentry.EventID

	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		scope.SetTag("error_source", source)
		scope.SetTag("status_code", fmt.Sprintf("%d", code))
		scope.SetTag("error_type", getErrorType(err))

		// Set request context
		scope.SetContext("request", map[string]any{
			"url":         c.OriginalURL(),
			"method":      c.Method(),
			"path":        c.Path(),
			"route":       c.Route().Path,
			"headers":     c.GetReqHeaders(),
			"user_agent":  c.Get("User-Agent"),
			"ip":          c.IP(),
			"body_size":   len(c.Body()),
			"query":       c.Queries(),
			"route_params": c.AllParams(),
		})

		// Set error details context
		scope.SetContext("error_details", map[string]any{
			"message":     err.Error(),
			"type":        fmt.Sprintf("%T", err),
			"stack_trace": string(debug.Stack()),
		})

		// Handle custom internal errors (erri)
		var internalErr *erri.Erri
		if errors.As(err, &internalErr) {
			scope.SetContext("internal_error", map[string]any{
				"type":         string(internalErr.Type),
				"message":      internalErr.Message,
				"details":      internalErr.Details,
				"property":     internalErr.Property,
				"value":        internalErr.Value,
				"file":         internalErr.File,
				"system_error": internalErr.SystemError,
			})

			scope.SetTag("internal_error_type", string(internalErr.Type))
			if internalErr.Property != "" {
				scope.SetTag("error_property", internalErr.Property)
			}

			// Better fingerprinting for grouping similar errors
			scope.SetFingerprint([]string{
				source,
				string(internalErr.Type),
				internalErr.Property,
			})
		} else {
			// Standard fingerprinting
			scope.SetFingerprint([]string{
				source,
				fmt.Sprintf("%d", code),
				getErrorFingerprint(err),
			})
		}

		// Add breadcrumb for the error
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Type:     "error",
			Category: source,
			Message:  err.Error(),
			Level:    sentry.LevelError,
			Data: map[string]any{
				"error_type": getErrorType(err),
				"status_code": code,
			},
		}, nil)

		eventID = hub.CaptureException(err)
	})

	return eventID
}

// EnhanceSentryEvent is deprecated - use ContextEnrichmentMiddleware from middleware.go
// Kept for backwards compatibility
func EnhanceSentryEvent(ctx *fiber.Ctx) error {
	if hub := sentryfiber.GetHubFromContext(ctx); hub != nil {
		hub.Scope().SetTag("endpoint", ctx.Route().Path)
		hub.Scope().SetTag("method", ctx.Method())

		if user := ctx.Locals("user"); user != nil {
			if userProvider, ok := user.(userIDProvider); ok {
				if userID := userProvider.GetUserID(); userID != "" {
					hub.Scope().SetUser(sentry.User{
						ID: userID,
					})
				}
			}
		}
	}
	return ctx.Next()
}

// RecoverMiddleware recovers from panics and sends them to Sentry
// This should be placed after the Sentry middleware but before other middleware
func RecoverMiddleware(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			hub := sentryfiber.GetHubFromContext(c)
			stackTrace := string(debug.Stack())
			errorLoc, file, line := core.ExtractErrorLocationWithDetails(stackTrace)

			var eventID *sentry.EventID
			if hub != nil {
				hub.WithScope(func(scope *sentry.Scope) {
					scope.SetLevel(sentry.LevelFatal)
					scope.SetTag("panic_recovered", "true")
					scope.SetTag("error_source", "panic_recovery")

					// Set panic details
					scope.SetContext("panic_details", map[string]any{
						"recovered_value": fmt.Sprintf("%v", r),
						"stack_trace":     stackTrace,
						"error_location":  errorLoc,
					})

					// Set request context
					scope.SetContext("request", map[string]any{
						"url":         c.OriginalURL(),
						"method":      c.Method(),
						"path":        c.Path(),
						"route":       c.Route().Path,
						"ip":          c.IP(),
						"user_agent":  c.Get("User-Agent"),
						"body_size":   len(c.Body()),
						"headers":     c.GetReqHeaders(),
					})

					// Set source location if available
					if file != "" && line > 0 {
						scope.SetTag("panic_file", file)
						scope.SetTag("panic_line", fmt.Sprintf("%d", line))
						scope.SetContext("source", map[string]any{
							"file": file,
							"line": line,
						})
					}

					// Add breadcrumb for the panic
					hub.AddBreadcrumb(&sentry.Breadcrumb{
						Type:     "error",
						Category: "panic",
						Message:  fmt.Sprintf("Panic recovered: %v", r),
						Level:    sentry.LevelFatal,
						Data: map[string]any{
							"recovered_value": fmt.Sprintf("%v", r),
							"location":        errorLoc,
						},
					}, nil)

					// Use RecoverWithContext for proper panic handling
					eventID = hub.RecoverWithContext(c.UserContext(), r)
				})
			}

			// Log the panic
			logFields := []any{
				slog.String("url", c.OriginalURL()),
				slog.String("method", c.Method()),
				slog.Any("panic_value", r),
				slog.String("error_location", errorLoc),
				slog.String("stack_trace", stackTrace),
			}

			if eventID != nil {
				logFields = append(logFields, slog.String("sentry_event_id", string(*eventID)))
			}

			if file != "" && line > 0 {
				logFields = append(logFields, slog.Any("source", slog.Source{
					File: file,
					Line: line,
				}))
			}

			handler.Log.ErrorContext(c.UserContext(), "Panic recovered in HTTP handler", logFields...)

			// Send 500 response
			_ = c.Status(http.StatusInternalServerError).SendString("Internal Server Error")
		}
	}()

	return c.Next()
}

// ErrorHandler is the global error handler for Fiber application
// It handles errors returned by route handlers and middleware
func ErrorHandler(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	// Determine status code
	code := fiber.StatusInternalServerError
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
	}

	// Check if it's a custom internal error
	var internalErr *erri.Erri
	if errors.As(err, &internalErr) {
		code = internalErr.HTTPStatusCode()
	}

	// Only send 5xx errors to Sentry
	if code >= 500 {
		eventID := captureErrorToSentry(c, err, code, "error_handler")

		// Log the error
		logFields := []any{
			slog.String("url", c.OriginalURL()),
			slog.String("method", c.Method()),
			slog.String("route", c.Route().Path),
			slog.Int("status_code", code),
			slog.Any("error", err),
			slog.String("error_type", getErrorType(err)),
		}

		if eventID != nil {
			logFields = append(logFields, slog.String("sentry_event_id", string(*eventID)))
		}

		// Add internal error details to log if available
		if internalErr != nil {
			logFields = append(logFields,
				slog.String("internal_error_type", string(internalErr.Type)),
				slog.String("internal_error_details", internalErr.Details),
				slog.String("internal_error_file", internalErr.File),
			)
		}

		handler.Log.ErrorContext(c.UserContext(), "Error handler: server error", logFields...)
	} else if code >= 400 {
		// Log 4xx errors as warnings (client errors, not our fault)
		handler.Log.WarnContext(c.UserContext(), "Error handler: client error",
			slog.String("url", c.OriginalURL()),
			slog.String("method", c.Method()),
			slog.Int("status_code", code),
			slog.String("error", err.Error()),
		)
	}

	return c.SendStatus(code)
}

// CaptureErrorMiddleware captures errors returned by handlers/middleware
// This is deprecated in favor of using ErrorHandler as the global error handler
// Kept for backwards compatibility
func CaptureErrorMiddleware(c *fiber.Ctx) error {
	err := c.Next()

	if err == nil {
		return nil
	}

	// Determine status code
	code := fiber.StatusInternalServerError
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
	}

	var internalErr *erri.Erri
	if errors.As(err, &internalErr) {
		code = internalErr.HTTPStatusCode()
	}

	// Only capture 5xx errors to Sentry
	if code >= 500 {
		eventID := captureErrorToSentry(c, err, code, "middleware")

		logFields := []any{
			slog.String("url", c.OriginalURL()),
			slog.String("method", c.Method()),
			slog.String("route", c.Route().Path),
			slog.Int("status_code", code),
			slog.Any("error", err),
			slog.String("error_type", getErrorType(err)),
		}

		if eventID != nil {
			logFields = append(logFields, slog.String("sentry_event_id", string(*eventID)))
		}

		if internalErr != nil {
			logFields = append(logFields,
				slog.String("internal_error_type", string(internalErr.Type)),
				slog.String("internal_error_details", internalErr.Details),
			)
		}

		handler.Log.ErrorContext(c.UserContext(), "Error captured in middleware", logFields...)
	}

	return err
}
