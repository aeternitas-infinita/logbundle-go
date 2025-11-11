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

// Log is the main logger instance for user code (always with source info)
var Log = slog.New(handler.NewCustomHandler(
	os.Stdout,
	core.GetLvlFromEnv("log_level"),
	true, // User logger always includes source
))

// InitLog reinitializes the global Log variable with the provided configuration
func InitLog(cfg LoggerConfig) {
	Log = CreateLogger(cfg)
}

// CreateLogger creates a new logger instance with the provided configuration
// This function is kept for future extensibility and user customization
func CreateLogger(loggerConfig LoggerConfig) *slog.Logger {
	h := handler.NewCustomHandler(os.Stdout, loggerConfig.Level, loggerConfig.AddSource)
	return slog.New(h)
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
