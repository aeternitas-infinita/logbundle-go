package lgsentry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
)

type Config struct {
	FilterLevels  []slog.Level
	ClientOptions sentry.ClientOptions
}

type integration struct {
	config    *Config
	initiated bool
}

var globalIntegration = &integration{}

func CaptureEvent(ctx context.Context, r slog.Record, args []slog.Attr) {
	config := globalIntegration.config

	shouldCapture := false

	if globalIntegration.initiated == false {
		return
	}

	for _, level := range config.FilterLevels {
		if r.Level >= level {
			shouldCapture = true
			break
		}
	}
	if shouldCapture == false {
		return
	}

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

	if traceID := core.GetTraceID(ctx); traceID != "" {
		tags[core.TraceIDKey] = traceID
	}

	tags["log_level"] = r.Level.String()
	extra["timestamp"] = r.Time.Format(time.RFC3339)

	if sourceInfo := extractSourceInfo(r); sourceInfo != nil {
		tags["source"] = fmt.Sprintf("%s:%d", sourceInfo.File, sourceInfo.Line)
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentryLevel)

		for key, value := range tags {
			scope.SetTag(key, value)
		}

		for key, value := range extra {
			scope.SetExtra(key, value)
		}

		scope.SetContext("log_context", map[string]any{
			"message":   r.Message,
			"level":     r.Level.String(),
			"timestamp": r.Time.Format(time.RFC3339),
			"source":    extra["source"],
		})

		if errorValue != nil {
			scope.SetTag("error_captured", "true")
			sentry.CaptureException(errorValue)
		} else {
			sentry.CaptureMessage(r.Message)
		}
	})
}

func Init(config *Config) error {
	globalIntegration = &integration{
		config: config,
	}

	sentryConfig := config.ClientOptions
	sentryConfig.BeforeSend = func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
		event.Tags["go_package"] = "log-bundle"
		return event
	}

	err := sentry.Init(sentryConfig)

	if err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	globalIntegration.initiated = true

	return nil
}

func Flush(timeout time.Duration) {
	if globalIntegration.initiated == false {
		sentry.Flush(timeout)
	}
}
