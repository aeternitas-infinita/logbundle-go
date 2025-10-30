package lgfiber

import (
	"context"
	"fmt"
	"time"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// BreadcrumbsMiddleware automatically adds breadcrumbs for each request
// This helps track the path leading to an error
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

		// Continue with the request
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
// This middleware should be placed after the Sentry middleware
func ContextEnrichmentMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		hub := sentryfiber.GetHubFromContext(c)
		if hub == nil {
			return c.Next()
		}

		// Set request tags
		hub.Scope().SetTag("http.method", c.Method())
		hub.Scope().SetTag("http.route", c.Route().Path)
		hub.Scope().SetTag("http.host", c.Hostname())
		hub.Scope().SetTag("http.protocol", c.Protocol())

		// Set request context with detailed information
		hub.Scope().SetContext("request", map[string]any{
			"url":          c.OriginalURL(),
			"method":       c.Method(),
			"path":         c.Path(),
			"route":        c.Route().Path,
			"ip":           c.IP(),
			"user_agent":   c.Get("User-Agent"),
			"content_type": c.Get("Content-Type"),
			"accept":       c.Get("Accept"),
			"referer":      c.Get("Referer"),
			"protocol":     c.Protocol(),
			"hostname":     c.Hostname(),
			"body_size":    len(c.Body()),
			"is_xhr":       c.XHR(),
			"is_secure":    c.Secure(),
			"remote_ip":    c.IP(),
			"ips":          c.IPs(),
		})

		// Add query parameters as context
		if queries := c.Queries(); len(queries) > 0 {
			queryParams := make(map[string]any)
			for k, v := range queries {
				queryParams[k] = v
			}
			hub.Scope().SetContext("query_params", queryParams)
		}

		// Add route parameters as context
		if params := c.AllParams(); len(params) > 0 {
			paramsAny := make(map[string]any)
			for k, v := range params {
				paramsAny[k] = v
			}
			hub.Scope().SetContext("route_params", paramsAny)
		}

		// Set user info if available
		if user := c.Locals("user"); user != nil {
			if userProvider, ok := user.(userIDProvider); ok {
				if userID := userProvider.GetUserID(); userID != "" {
					hub.Scope().SetUser(sentry.User{
						ID: userID,
					})
					hub.Scope().SetTag("user.id", userID)
				}
			}
		}

		// Store fiber context in Go context for access in slog handlers
		ctx := context.WithValue(c.UserContext(), "fiber_ctx", c)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// PerformanceMiddleware tracks request performance with Sentry transactions
// This requires EnablePerformance to be true in Sentry config
func PerformanceMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		hub := sentryfiber.GetHubFromContext(c)
		if hub == nil {
			return c.Next()
		}

		// Start a transaction for this request
		transactionName := fmt.Sprintf("%s %s", c.Method(), c.Route().Path)

		// Create transaction context
		ctx := c.UserContext()
		transaction := sentry.StartTransaction(
			ctx,
			transactionName,
			sentry.WithOpName("http.server"),
			sentry.WithTransactionSource(sentry.SourceRoute),
		)
		defer transaction.Finish()

		// Set transaction on the scope
		hub.Scope().SetContext("trace", map[string]any{
			"trace_id":       transaction.TraceID.String(),
			"span_id":        transaction.SpanID.String(),
			"parent_span_id": transaction.ParentSpanID.String(),
		})

		// Add transaction data
		transaction.SetData("http.method", c.Method())
		transaction.SetData("http.route", c.Route().Path)
		transaction.SetData("http.url", c.OriginalURL())

		// Store transaction in context
		ctx = transaction.Context()
		c.SetUserContext(ctx)

		// Execute the request
		err := c.Next()

		// Set transaction status based on response
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

// StartSpan starts a new span for tracking operations within a request
// Usage: defer lgfiber.StartSpan(c, "operation.name", "description").Finish()
func StartSpan(c *fiber.Ctx, operation, description string) *sentry.Span {
	ctx := c.UserContext()
	span := sentry.StartSpan(ctx, operation)
	span.Description = description

	// Update context with new span
	c.SetUserContext(span.Context())

	return span
}

// AddBreadcrumb is a helper to add custom breadcrumbs from handlers
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

// SetTag is a helper to set tags on the current scope
func SetTag(c *fiber.Ctx, key, value string) {
	hub := sentryfiber.GetHubFromContext(c)
	if hub == nil {
		return
	}
	hub.Scope().SetTag(key, value)
}

// SetContext is a helper to set context data on the current scope
func SetContext(c *fiber.Ctx, key string, value map[string]any) {
	hub := sentryfiber.GetHubFromContext(c)
	if hub == nil {
		return
	}
	hub.Scope().SetContext(key, value)
}

// TraceIDMiddleware extracts Sentry's trace_id and injects it into the request context
// This trace_id will be automatically included in all logs for this request
// Must be placed AFTER Sentry middleware and AFTER PerformanceMiddleware
func TraceIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		hub := sentryfiber.GetHubFromContext(c)
		if hub == nil {
			// Fallback to UUID if Sentry is not available
			traceID := uuid.New().String()
			ctx := context.WithValue(c.UserContext(), core.TraceIDKey, traceID)
			c.SetUserContext(ctx)
			return c.Next()
		}

		// Get the transaction from Sentry span
		span := sentry.SpanFromContext(c.UserContext())
		var traceID string

		if span != nil {
			// Use Sentry's trace ID from the span/transaction
			traceID = span.TraceID.String()
		} else {
			// Fallback to UUID if no transaction exists
			traceID = uuid.New().String()
		}

		// Add trace_id to context for logging
		ctx := context.WithValue(c.UserContext(), core.TraceIDKey, traceID)
		c.SetUserContext(ctx)

		// Also add to Sentry tags for easy filtering
		hub.Scope().SetTag(core.TraceIDKey, traceID)

		return c.Next()
	}
}
