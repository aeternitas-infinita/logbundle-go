package lgsentry

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/config"
)

func CaptureEvent(ctx context.Context, level sentry.Level, msg string, err error, extraData ...any) {
	// Check if Sentry is globally enabled
	if !config.IsSentryEnabled() {
		return
	}

	// Check context cancellation before expensive operations
	if ctx != nil {
		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	var hub *sentry.Hub
	var fiberCtx *fiber.Ctx

	if ctx != nil {
		if fc, ok := ctx.Value("fiber_ctx").(*fiber.Ctx); ok && fc != nil {
			fiberCtx = fc
			hub = sentryfiber.GetHubFromContext(fc)
		}
	}

	if hub == nil {
		hub = sentry.CurrentHub()
	}

	tags, extra := parseExtraData(extraData)

	captureFunc := func(scope *sentry.Scope) {
		scope.SetLevel(level)

		for key, value := range tags {
			scope.SetTag(key, value)
		}

		for key, value := range extra {
			scope.SetExtra(key, value)
		}

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
				scope.SetExtra("query_params", queries)
			}
			if params := fiberCtx.AllParams(); len(params) > 0 {
				scope.SetExtra("route_params", params)
			}
		}

		if err != nil {
			scope.SetContext("error_details", map[string]any{
				"message": msg,
				"error":   err.Error(),
			})

			captureErr := fmt.Errorf("%s: %w", msg, err)
			hub.CaptureException(captureErr)
		} else {
			scope.SetContext("log_context", map[string]any{
				"message": msg,
			})
			hub.CaptureMessage(msg)
		}
	}

	hub.WithScope(captureFunc)
}

func parseExtraData(extraData []any) (map[string]string, map[string]any) {
	tags := make(map[string]string)
	extra := make(map[string]any)

	const maxTagLength = 100

	for i := 0; i < len(extraData); i++ {
		if attr, ok := extraData[i].(slog.Attr); ok {
			key := attr.Key
			value := attr.Value.Any()

			if _, isErr := value.(error); isErr {
				continue
			}

			if strVal, ok := value.(string); ok {
				if len(strVal) < maxTagLength && !strings.Contains(strVal, "\n") {
					tags[key] = strVal
					continue
				}
			}

			switch v := value.(type) {
			case int:
				tags[key] = fmt.Sprintf("%d", v)
			case int64:
				tags[key] = fmt.Sprintf("%d", v)
			case float64:
				tags[key] = fmt.Sprintf("%f", v)
			case bool:
				tags[key] = fmt.Sprintf("%t", v)
			default:
				extra[key] = value
			}
		}
	}

	return tags, extra
}
