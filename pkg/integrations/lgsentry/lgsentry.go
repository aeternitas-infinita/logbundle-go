// Package lgsentry provides Sentry integration for slog-based logging.
// It automatically captures log events at specified levels and sends them to Sentry
// with full context enrichment from Fiber requests.
package lgsentry

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gofiber/fiber/v2"
)

// integration holds the global Sentry integration state
type integration struct {
	config    *Config
	initiated bool
}

var globalIntegration = &integration{}

// Init initializes the Sentry client with the provided configuration
// Must be called before any logging with Sentry enabled
func Init(config *Config) error {
	globalIntegration = &integration{
		config: config,
	}

	// Wrap user's BeforeSend callback with additional enrichment
	userBeforeSend := config.BeforeSend
	config.BeforeSend = func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
		// Initialize tags if needed
		if event.Tags == nil {
			event.Tags = make(map[string]string)
		}

		// Enhance event with Fiber request data if available
		if hint != nil && hint.Context != nil {
			if fc, ok := hint.Context.Value("fiber_ctx").(*fiber.Ctx); ok && fc != nil {
				// Initialize extra data if needed
				if event.Extra == nil {
					event.Extra = make(map[string]any)
				}

				// Add query parameters
				if queries := fc.Queries(); len(queries) > 0 {
					event.Extra["query_params"] = queries
				}

				// Add route parameters
				if params := fc.AllParams(); len(params) > 0 {
					event.Extra["route_params"] = params
				}

				// Add detailed request information
				if event.Contexts == nil {
					event.Contexts = make(map[string]sentry.Context)
				}
				event.Contexts["custom_request"] = map[string]any{
					"content_length": fc.Context().Request.Header.ContentLength(),
					"protocol":       string(fc.Context().Request.Header.Protocol()),
					"host":           fc.Hostname(),
					"is_tls":         fc.Protocol() == "https",
				}
			}
		}

		// Call user-provided BeforeSend if exists
		if userBeforeSend != nil {
			return userBeforeSend(event, hint)
		}
		return event
	}

	// Initialize Sentry SDK with the config (which embeds ClientOptions)
	if err := sentry.Init(config.ClientOptions); err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	globalIntegration.initiated = true

	return nil
}

// Flush waits up to the given timeout for all events to be sent to Sentry
// Should be called before application shutdown
func Flush(timeout time.Duration) {
	if globalIntegration.initiated {
		sentry.Flush(timeout)
	}
}
