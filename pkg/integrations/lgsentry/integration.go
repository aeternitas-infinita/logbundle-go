// rev.

package lgsentry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"
	sentryfiber "github.com/getsentry/sentry-go/fiber"
	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
)

// Config extends sentry.ClientOptions with additional logbundle-specific settings.
type Config struct {
	sentry.ClientOptions

	// For example: []slog.Level{slog.LevelWarn, slog.LevelError}
	// Only logs at these levels or higher will be captured.
	FilterLevels []slog.Level
}

type integration struct {
	config    *Config
	initiated bool
}

var globalIntegration = &integration{}

func CaptureEventForSlog(ctx context.Context, r slog.Record, args []slog.Attr) {
	config := globalIntegration.config

	// check if Sentry is initiated
	if globalIntegration.initiated == false {
		return
	}

	// check if level should be captured
	shouldCapture := false
	for _, level := range config.FilterLevels {
		if r.Level >= level {
			shouldCapture = true
			break
		}
	}
	if !shouldCapture {
		return
	}

	// convert slog level to Sentry level
	var sentryLevel sentry.Level
	switch r.Level {
	case slog.LevelDebug:
		sentryLevel = sentry.LevelDebug
	case slog.LevelInfo:
		sentryLevel = sentry.LevelInfo
	case slog.LevelWarn:
		sentryLevel = sentry.LevelWarning
	case slog.LevelError:
		sentryLevel = sentry.LevelError
	default:
		sentryLevel = sentry.LevelInfo
	}

	tags, extra, errorValue := extractSentryData(args)

	// add log trace ID if available
	if traceID := core.GetLogTraceID(ctx); traceID != "" {
		tags[core.TraceIDKey] = traceID
	}

	tags["log_level"] = r.Level.String()
	extra["timestamp"] = r.Time.Format(time.RFC3339)

	// add source info
	if sourceInfo := extractSourceInfo(r); sourceInfo != nil {
		tags["source"] = fmt.Sprintf("%s:%d", sourceInfo.File, sourceInfo.Line)
		extra["source_file"] = sourceInfo.File
		extra["source_line"] = sourceInfo.Line
	}

	// try to get Fiber context and Hub from context
	var hub *sentry.Hub
	var fiberCtx *fiber.Ctx

	// extract fiber.Ctx if it exists in context
	if ctx != nil {
		if fc, ok := ctx.Value("fiber_ctx").(*fiber.Ctx); ok {
			fiberCtx = fc
			hub = sentryfiber.GetHubFromContext(fc)
		}
	}

	// capture function that works with both hub and global scope
	captureFunc := func(scope *sentry.Scope) {
		scope.SetLevel(sentryLevel)

		// set tags
		for key, value := range tags {
			scope.SetTag(key, value)
		}

		// set extra data
		for key, value := range extra {
			scope.SetExtra(key, value)
		}

		// set log context
		scope.SetContext("log_context", map[string]any{
			"message":   r.Message,
			"level":     r.Level.String(),
			"timestamp": r.Time.Format(time.RFC3339),
			"source":    tags["source"],
		})

		// add request context if Fiber context is available
		if fiberCtx != nil {
			scope.SetContext("request", map[string]any{
				"url":        fiberCtx.OriginalURL(),
				"method":     fiberCtx.Method(),
				"path":       fiberCtx.Path(),
				"route":      fiberCtx.Route().Path,
				"ip":         fiberCtx.IP(),
				"user_agent": fiberCtx.Get("User-Agent"),
			})

			// add breadcrumb for this log entry
			hub.AddBreadcrumb(&sentry.Breadcrumb{
				Category:  "log",
				Message:   r.Message,
				Level:     sentryLevel,
				Timestamp: r.Time,
				Data:      extra,
			}, nil)
		}

		// capture the event
		if errorValue != nil {
			scope.SetTag("error_captured", "true")
			if hub != nil {
				hub.CaptureException(errorValue)
			} else {
				sentry.CaptureException(errorValue)
			}
		} else {
			if hub != nil {
				hub.CaptureMessage(r.Message)
			} else {
				sentry.CaptureMessage(r.Message)
			}
		}
	}

	// use hub if available, otherwise use global scope
	if hub != nil {
		hub.WithScope(captureFunc)
	} else {
		sentry.WithScope(captureFunc)
	}
}

func Init(config *Config) error {
	globalIntegration = &integration{
		config: config,
	}

	// wrap user's BeforeSend with our custom logic
	userBeforeSend := config.BeforeSend
	config.BeforeSend = func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
		// add default tags
		if event.Tags == nil {
			event.Tags = make(map[string]string)
		}

		// enhance event with Fiber request data if available
		if hint != nil && hint.Context != nil {
			if fc, ok := hint.Context.Value("fiber_ctx").(*fiber.Ctx); ok && fc != nil {
				// initialize Extra if nil
				if event.Extra == nil {
					event.Extra = make(map[string]any)
				}

				// add query parameters
				if queries := fc.Queries(); len(queries) > 0 {
					event.Extra["query_params"] = queries
				}

				// add route parameters
				if params := fc.AllParams(); len(params) > 0 {
					event.Extra["route_params"] = params
				}

				// add custom request info
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

		// call user-provided BeforeSend if exists
		if userBeforeSend != nil {
			return userBeforeSend(event, hint)
		}
		return event
	}

	// initialize Sentry with the config (which embeds ClientOptions)
	err := sentry.Init(config.ClientOptions)
	if err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	globalIntegration.initiated = true

	return nil
}

func Flush(timeout time.Duration) {
	if globalIntegration.initiated {
		sentry.Flush(timeout)
	}
}
