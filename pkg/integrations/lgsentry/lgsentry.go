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

		// Improve error message for better Sentry issue titles
		if len(event.Exception) > 0 {
			exception := event.Exception[0]

			// List of generic Go error types that need better messages
			genericErrorTypes := []string{
				"*fmt.wrapError",
				"*errors.errorString",
				"*errors.joinError",
				"*url.Error",
				"errors.errorString",
				"fmt.wrapError",
			}

			// Check if this is a generic error type
			isGenericType := false
			for _, genericType := range genericErrorTypes {
				if exception.Type == genericType {
					isGenericType = true
					break
				}
			}

			// Extract a better message
			var betterMessage string

			// Priority 1: Use Message from internal_error context (for erri.Erri)
			if event.Contexts != nil {
				if internalErr, ok := event.Contexts["internal_error"]; ok {
					if msg, ok := internalErr["message"].(string); ok && msg != "" {
						betterMessage = msg
					} else if details, ok := internalErr["details"].(string); ok && details != "" {
						betterMessage = details
					}
				}

				// Priority 2: Use message from error_details context
				if betterMessage == "" {
					if errorDetails, ok := event.Contexts["error_details"]; ok {
						if msg, ok := errorDetails["message"].(string); ok && msg != "" {
							betterMessage = msg
						}
					}
				}
			}

			// Priority 3: Use exception value if available
			if betterMessage == "" && exception.Value != "" {
				betterMessage = exception.Value
			}

			// Apply the better message if found
			if betterMessage != "" {
				// Truncate long messages for readability (max 200 chars)
				if len(betterMessage) > 200 {
					betterMessage = betterMessage[:197] + "..."
				}

				// Update the exception value (this becomes the Sentry issue title)
				event.Exception[0].Value = betterMessage

				// Also update the event message for consistency
				if event.Message == "" || isGenericType {
					event.Message = betterMessage
				}
			}
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
