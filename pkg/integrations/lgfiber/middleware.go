package lgfiber

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/config"
	"github.com/aeternitas-infinita/logbundle-go/pkg/handler"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"
)

func BreadcrumbsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip breadcrumbs if Sentry disabled to avoid allocations
		if !config.IsSentryEnabled() {
			return c.Next()
		}

		hub := sentryfiber.GetHubFromContext(c)
		if hub == nil {
			return c.Next()
		}

		startTime := time.Now()

		// Add request start breadcrumb
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Type:      "http",
			Category:  "request.start",
			Message:   fmt.Sprintf("%s %s", c.Method(), c.Path()),
			Level:     sentry.LevelInfo,
			Timestamp: startTime,
			Data: map[string]any{
				"url":    c.OriginalURL(),
				"method": c.Method(),
				"path":   c.Path(),
				"route":  c.Route().Path,
				"ip":     c.IP(),
			},
		}, nil)

		err := c.Next()

		// Add request end breadcrumb
		duration := time.Since(startTime)
		statusCode := c.Response().StatusCode()

		breadcrumbLevel := sentry.LevelInfo
		if statusCode >= 500 {
			breadcrumbLevel = sentry.LevelError
		} else if statusCode >= 400 {
			breadcrumbLevel = sentry.LevelWarning
		}

		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Type:      "http",
			Category:  "request.end",
			Message:   fmt.Sprintf("%s %s - %d", c.Method(), c.Path(), statusCode),
			Level:     breadcrumbLevel,
			Timestamp: time.Now(),
			Data: map[string]any{
				"status_code":   statusCode,
				"duration_ms":   duration.Milliseconds(),
				"response_size": len(c.Response().Body()),
			},
		}, nil)

		return err
	}
}

func RecoverMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// Use middleware logger if configured, otherwise fall back to internal logger
				log := config.GetMiddlewareLogger()
				if log == nil {
					log = handler.GetInternalLogger()
				}

				log.Error("Panic recovered",
					slog.String("panic", fmt.Sprintf("%v", r)),
					slog.String("url", c.OriginalURL()),
					slog.String("method", c.Method()),
				)

				c.Status(fiber.StatusInternalServerError).JSON(lgerr.ErrorResponse{
					Title:  "Internal Server Error",
					Detail: "An unexpected error occurred",
				})
			}
		}()

		return c.Next()
	}
}

// StartSpan starts a new Sentry span for the current request
func StartSpan(c *fiber.Ctx, operation, description string) *sentry.Span {
	ctx := c.UserContext()
	span := sentry.StartSpan(ctx, operation)
	span.Description = description
	c.SetUserContext(span.Context())
	return span
}

// AddBreadcrumb adds a custom breadcrumb to Sentry
func AddBreadcrumb(c *fiber.Ctx, category, message string, level sentry.Level, data map[string]any) {
	hub := sentryfiber.GetHubFromContext(c)
	if hub == nil {
		return
	}

	hub.AddBreadcrumb(&sentry.Breadcrumb{
		Category:  category,
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
		Data:      data,
	}, nil)
}

// SetTag sets a tag on the current Sentry scope
func SetTag(c *fiber.Ctx, key, value string) {
	hub := sentryfiber.GetHubFromContext(c)
	if hub == nil {
		return
	}
	hub.Scope().SetTag(key, value)
}

// SetContext sets context data on the current Sentry scope
func SetContext(c *fiber.Ctx, key string, value map[string]any) {
	hub := sentryfiber.GetHubFromContext(c)
	if hub == nil {
		return
	}
	hub.Scope().SetContext(key, value)
}
