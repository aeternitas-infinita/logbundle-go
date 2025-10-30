package lgfiber

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/erri"
)

// captureErrorToSentry is an internal helper for capturing errors to Sentry
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
			"url":          c.OriginalURL(),
			"method":       c.Method(),
			"path":         c.Path(),
			"route":        c.Route().Path,
			"headers":      c.GetReqHeaders(),
			"user_agent":   c.Get("User-Agent"),
			"ip":           c.IP(),
			"body_size":    len(c.Body()),
			"query":        c.Queries(),
			"route_params": c.AllParams(),
		})

		// Set error details context
		scope.SetContext("error_details", map[string]any{
			"message": err.Error(),
			"type":    fmt.Sprintf("%T", err),
		})

		// Handle custom internal errors (erri)
		var internalErr *erri.Erri
		if errors := err; errors != nil {
			if ie, ok := errors.(*erri.Erri); ok {
				internalErr = ie
			}
		}

		if internalErr != nil {
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
				"error_type":  getErrorType(err),
				"status_code": code,
			},
		}, nil)

		eventID = hub.CaptureException(err)
	})

	return eventID
}
