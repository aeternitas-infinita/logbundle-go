package logbundle

import (
	"log/slog"
	"os"

	"github.com/aeternitas-infinita/logbundle-go/pkg/config"
	"github.com/aeternitas-infinita/logbundle-go/pkg/handler"
)

// LoggerConfig holds configuration options for creating a logger instance
type LoggerConfig struct {
	Level     slog.Level // Minimum log level to output (Debug, Info, Warn, Error)
	AddSource bool       // Whether to include source file and line number in logs
}

// CreateLogger creates a new logger instance with the provided configuration
// If setAsMiddlewareLogger is true, this logger will be used by all middlewares
func CreateLogger(loggerConfig LoggerConfig, setAsMiddlewareLogger ...bool) *slog.Logger {
	h := handler.NewCustomHandler(os.Stdout, loggerConfig.Level, loggerConfig.AddSource)
	logger := slog.New(h)

	// If setAsMiddlewareLogger is true, set this logger for middleware use
	if len(setAsMiddlewareLogger) > 0 && setAsMiddlewareLogger[0] {
		config.SetMiddlewareLogger(logger)
	}

	return logger
}

// SetMiddlewareLogger sets the logger to be used by all middlewares
// If not set, middlewares will use the internal logger
func SetMiddlewareLogger(logger *slog.Logger) {
	config.SetMiddlewareLogger(logger)
}

// GetMiddlewareLogger returns the configured middleware logger, or nil if not set
func GetMiddlewareLogger() *slog.Logger {
	return config.GetMiddlewareLogger()
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
