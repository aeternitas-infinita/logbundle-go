package lgfiber

import (
	"context"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
)

// userIDProvider interface for extracting user ID from context locals
type userIDProvider interface {
	GetUserID() string
}

// BreadcrumbsMiddleware adds HTTP request breadcrumbs to Sentry
func BreadcrumbsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
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

// ContextEnrichmentMiddleware enriches Sentry context with request data
func ContextEnrichmentMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		hub := sentryfiber.GetHubFromContext(c)
		if hub == nil {
			return c.Next()
		}

		// Set basic HTTP tags
		hub.Scope().SetTag("http.method", c.Method())
		hub.Scope().SetTag("http.route", c.Route().Path)
		hub.Scope().SetTag("http.host", c.Hostname())

		// Set request context
		hub.Scope().SetContext("request", map[string]any{
			"url":          c.OriginalURL(),
			"method":       c.Method(),
			"path":         c.Path(),
			"route":        c.Route().Path,
			"ip":           c.IP(),
			"user_agent":   c.Get("User-Agent"),
			"content_type": c.Get("Content-Type"),
			"protocol":     c.Protocol(),
			"hostname":     c.Hostname(),
		})

		// Add query params if present
		if queries := c.Queries(); len(queries) > 0 {
			// Pre-allocate map with known size
			queryParams := make(map[string]any, len(queries))
			for k, v := range queries {
				queryParams[k] = v
			}
			hub.Scope().SetContext("query_params", queryParams)
		}

		// Add route params if present
		if params := c.AllParams(); len(params) > 0 {
			// Pre-allocate map with known size
			routeParams := make(map[string]any, len(params))
			for k, v := range params {
				routeParams[k] = v
			}
			hub.Scope().SetContext("route_params", routeParams)
		}

		// Add user info if available
		if user := c.Locals("user"); user != nil {
			if userProvider, ok := user.(userIDProvider); ok {
				if userID := userProvider.GetUserID(); userID != "" {
					hub.Scope().SetUser(sentry.User{ID: userID})
					hub.Scope().SetTag("user.id", userID)
				}
			}
		}

		// Store fiber context in user context for later use
		// Using a private type as key to avoid collisions
		type fiberCtxKey struct{}
		ctx := context.WithValue(c.UserContext(), fiberCtxKey{}, c)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// PerformanceMiddleware creates Sentry performance transactions for requests
func PerformanceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		hub := sentryfiber.GetHubFromContext(c)
		if hub == nil {
			return c.Next()
		}

		transactionName := fmt.Sprintf("%s %s", c.Method(), c.Route().Path)
		ctx := c.UserContext()

		transaction := sentry.StartTransaction(
			ctx,
			transactionName,
			sentry.WithOpName("http.server"),
			sentry.WithTransactionSource(sentry.SourceRoute),
		)
		defer transaction.Finish()

		// Set transaction context
		hub.Scope().SetContext("trace", map[string]any{
			"trace_id": transaction.TraceID.String(),
			"span_id":  transaction.SpanID.String(),
		})

		transaction.SetData("http.method", c.Method())
		transaction.SetData("http.route", c.Route().Path)
		transaction.SetData("http.url", c.OriginalURL())

		ctx = transaction.Context()
		c.SetUserContext(ctx)

		err := c.Next()

		// Set transaction status based on HTTP status
		statusCode := c.Response().StatusCode()
		transaction.SetData("http.status_code", statusCode)

		if statusCode >= 500 {
			transaction.Status = sentry.SpanStatusInternalError
		} else if statusCode >= 400 {
			transaction.Status = sentry.SpanStatusInvalidArgument
		} else if statusCode >= 200 && statusCode < 300 {
			transaction.Status = sentry.SpanStatusOK
		}

		return err
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
