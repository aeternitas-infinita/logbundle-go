package lgsentry

import (
	"log/slog"

	"github.com/getsentry/sentry-go"
)

// shouldCaptureLevel checks if the given log level should be sent to Sentry
// If filterLevels is empty/nil, all logs are captured
// If filterLevels is set, only logs at those levels or higher are captured
func shouldCaptureLevel(level slog.Level, filterLevels []slog.Level) bool {
	if len(filterLevels) == 0 {
		return true
	}

	for _, filterLevel := range filterLevels {
		if level >= filterLevel {
			return true
		}
	}
	return false
}

// convertLogLevelToSentry maps slog levels to Sentry levels
func convertLogLevelToSentry(level slog.Level) sentry.Level {
	switch level {
	case slog.LevelDebug:
		return sentry.LevelDebug
	case slog.LevelInfo:
		return sentry.LevelInfo
	case slog.LevelWarn:
		return sentry.LevelWarning
	case slog.LevelError:
		return sentry.LevelError
	default:
		return sentry.LevelInfo
	}
}
