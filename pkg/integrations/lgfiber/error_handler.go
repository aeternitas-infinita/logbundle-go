package lgfiber

import (
	"context"
	"errors"

	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"
	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
)

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
			WithHTTPStatus(code)

		// Map common HTTP status codes to appropriate error types
		if code == fiber.StatusNotFound {
			lgErr.WithType(lgerr.TypeNotFound).WithTitle("Not Found")
		} else if code >= 500 {
			lgErr.WithTitle("Internal Server Error")
		} else if code >= 400 {
			lgErr.WithTitle("Bad Request")
		}

		// Continue with normal lgerr.Error handling flow
	}

	// Handle lgerr.Error
	var sentryEventID *sentry.EventID

	// Lightweight pre-check first
	if shouldSendToSentryLazy(lgErr) {
		// Only fetch hub if pre-check passed
		hub := sentryfiber.GetHubFromContext(c)
		if shouldSendToSentry(lgErr, hub) {
			sentryEventID = captureToSentry(c.UserContext(), hub, lgErr, "error_handler", c)
		}
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
