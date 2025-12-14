package logbundle

import (
	"log/slog"
	"os"

	"github.com/aeternitas-infinita/logbundle-go/pkg/config"
	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
	"github.com/aeternitas-infinita/logbundle-go/pkg/handler"
)

// LoggerConfig holds configuration options for creating a logger instance
type LoggerConfig struct {
	Level     slog.Level // Minimum log level to output (Debug, Info, Warn, Error)
	AddSource bool       // Whether to include source file and line number in logs
}

// CreateLogger creates a new logger instance with the provided configuration
func CreateLogger(loggerConfig LoggerConfig) *slog.Logger {
	h := handler.NewCustomHandler(os.Stdout, loggerConfig.Level, loggerConfig.AddSource)
	return slog.New(h)
}

// CreateLoggerDefault creates a logger with default configuration from environment
func CreateLoggerDefault() *slog.Logger {
	return CreateLogger(LoggerConfig{
		Level:     core.GetLvlFromEnv("log_level"),
		AddSource: true,
	})
}

// IsSentryEnabled returns whether Sentry integration is currently enabled
func IsSentryEnabled() bool {
	return config.IsSentryEnabled()
}

// SetSentryEnabled enables or disables Sentry integration globally
// When disabled, no events will be sent to Sentry from any part of the library
func SetSentryEnabled(enabled bool) {
	config.SetSentryEnabled(enabled)
}

// GetSentryMinHTTPStatus returns the minimum HTTP status code to send to Sentry
func GetSentryMinHTTPStatus() int {
	return config.GetSentryMinHTTPStatus()
}

// SetSentryMinHTTPStatus sets the minimum HTTP status code to send to Sentry
// Examples:
//   - 500: Only server errors (5xx) - default
//   - 400: Client and server errors (4xx and 5xx)
//   - 0: All errors regardless of status code
func SetSentryMinHTTPStatus(minStatus int) {
	config.SetSentryMinHTTPStatus(minStatus)
}
