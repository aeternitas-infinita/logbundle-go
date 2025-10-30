package lgsentry

import (
	"log/slog"

	"github.com/getsentry/sentry-go"
)

// Config extends sentry.ClientOptions with additional logbundle-specific settings
type Config struct {
	sentry.ClientOptions

	// FilterLevels specifies which slog levels should be sent to Sentry
	// Example: []slog.Level{slog.LevelWarn, slog.LevelError}
	// Only logs at these levels or higher will be captured
	FilterLevels []slog.Level
}
