package lgfiber

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

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

func RecoverMiddleware(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			hub := sentryfiber.GetHubFromContext(c)

			var eventID *sentry.EventID
			if hub != nil {
				hub.WithScope(func(scope *sentry.Scope) {
					scope.SetTag("panic_recovered", "true")
					scope.SetTag("component", "fiber_recovery_middleware")

					scope.SetContext("panic_details", map[string]any{
						"recovered_value": fmt.Sprintf("%v", r),
						"stack_trace":     string(debug.Stack()),
					})

					scope.AddEventProcessor(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
						if event != nil {
							event.Message = fmt.Sprintf("Panic: %v", r)
						}
						return event
					})

					eventID = hub.RecoverWithContext(c.Context(), r)
				})
			}

			stackTrace := string(debug.Stack())
			errorLoc := core.ExtractErrorLocation(stackTrace)

			logFields := []any{
				slog.String("url", c.OriginalURL()),
				slog.Any("error", r),
				slog.String("error_location", fmt.Sprintf("[%s]", errorLoc)),
			}

			if eventID != nil {
				logFields = append(logFields, slog.String("sentry_event_id", string(*eventID)))
			}

			handler.Log.ErrorContext(c.Context(), "Panic recovered", logFields...)

			if eventID != nil && hub != nil {
				hub.Flush(10 * time.Second)
			}

			c.SendStatus(http.StatusInternalServerError)
		}
	}()

	return c.Next()
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
	}

	if code >= 500 {
		if hub := sentryfiber.GetHubFromContext(c); hub != nil {
			hub.WithScope(func(scope *sentry.Scope) {
				scope.AddEventProcessor(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
					if event != nil {
						event.Message = "Error handler error"
					}
					return event
				})
				scope.SetLevel(sentry.LevelError)

				scope.SetTag("error_handler", "fiber")
				scope.SetTag("status_code", fmt.Sprintf("%d", code))
				scope.SetTag("error_type", getErrorType(err))

				scope.SetContext("request", map[string]any{
					"url":        c.OriginalURL(),
					"method":     c.Method(),
					"headers":    c.GetReqHeaders(),
					"user_agent": c.Get("User-Agent"),
					"ip":         c.IP(),
					"body_size":  len(c.Body()),
					"query":      c.Queries(),
				})

				scope.SetContext("error_details", map[string]any{
					"message":     err.Error(),
					"type":        fmt.Sprintf("%T", err),
					"stack_trace": string(debug.Stack()),
				})

				scope.SetFingerprint([]string{
					"fiber-error",
					fmt.Sprintf("%d", code),
					getErrorFingerprint(err),
				})

				eventID := hub.CaptureException(err)

				logFields := []any{
					slog.String("url", c.OriginalURL()),
					slog.String("method", c.Method()),
					slog.Int("status_code", code),
					slog.Any("error", err),
					slog.String("stack_trace", string(debug.Stack())),
				}

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
				}

				if eventID != nil {
					logFields = append(logFields, slog.String("sentry_event_id", string(*eventID)))
				}

				handler.Log.ErrorContext(c.Context(), "Error handler: handled server error", logFields...)
			})
		} else {
			handler.Log.ErrorContext(c.Context(), "Error handler: handled server error",
				slog.String("url", c.OriginalURL()),
				slog.String("method", c.Method()),
				slog.Int("status_code", code),
				slog.Any("error", err),
				slog.String("stack_trace", string(debug.Stack())),
			)
		}
	}

	return c.SendStatus(code)
}

func CaptureErrorMiddleware(c *fiber.Ctx) error {
	err := c.Next()

	if err != nil {
		code := fiber.StatusInternalServerError
		var e *fiber.Error
		if errors.As(err, &e) {
			code = e.Code
		}

		if code >= 500 {
			if hub := sentryfiber.GetHubFromContext(c); hub != nil {
				hub.WithScope(func(scope *sentry.Scope) {
					scope.AddEventProcessor(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
						if event != nil {
							event.Message = "Error captured in middleware"
						}
						return event
					})
					scope.SetLevel(sentry.LevelError)

					scope.SetTag("error_handler", "middleware")
					scope.SetTag("status_code", fmt.Sprintf("%d", code))
					scope.SetTag("error_type", getErrorType(err))

					scope.SetContext("request", map[string]any{
						"url":        c.OriginalURL(),
						"method":     c.Method(),
						"headers":    c.GetReqHeaders(),
						"user_agent": c.Get("User-Agent"),
						"ip":         c.IP(),
						"body_size":  len(c.Body()),
						"query":      c.Queries(),
					})

					scope.SetContext("error_details", map[string]any{
						"message":     err.Error(),
						"type":        fmt.Sprintf("%T", err),
						"stack_trace": string(debug.Stack()),
					})

					scope.SetFingerprint([]string{
						"middleware-error",
						fmt.Sprintf("%d", code),
						getErrorFingerprint(err),
					})

					logFields := []any{
						slog.String("url", c.OriginalURL()),
						slog.String("method", c.Method()),
						slog.Int("status_code", code),
						slog.Any("error", err),
						slog.String("stack_trace", string(debug.Stack())),
					}

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
					}

					eventID := hub.CaptureException(err)

					if eventID != nil {
						logFields = append(logFields, slog.String("sentry_event_id", string(*eventID)))
					}

					handler.Log.ErrorContext(c.Context(), "Error captured in middleware", logFields...)
				})
			} else {
				handler.Log.ErrorContext(c.Context(), "Error captured in middleware",
					slog.String("url", c.OriginalURL()),
					slog.String("method", c.Method()),
					slog.Int("status_code", code),
					slog.Any("error", err),
					slog.String("stack_trace", string(debug.Stack())),
				)
			}
		}
	}

	return err
}
