package logbundle

import (
	"log/slog"
	"os"

	"github.com/aeternitas-infinita/logbundle-go/internal/handler"
	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
)

// LoggerConfig defines the logger configuration
type LoggerConfig struct {
	Level         slog.Level
	SentryEnabled bool
	AddSource     bool
}

var (
	// Log is the default logger instance with source information
	Log = slog.New(handler.NewCustomHandler(
		os.Stdout,
		core.GetLvlFromEnv("log_level"),
		true,
		false,
	))

	// LogMin is the minimal logger instance without source information
	LogMin = slog.New(handler.NewCustomHandler(
		os.Stdout,
		core.GetLvlFromEnv("log_level"),
		false,
		false,
	))
)

// InitLog reinitializes the default logger with custom configuration
func InitLog(cfg LoggerConfig) {
	Log = CreateLogger(cfg)
}

// InitLogMin reinitializes the minimal logger with custom configuration
func InitLogMin(cfg LoggerConfig) {
	LogMin = CreateLogger(cfg)
}

// CreateLogger creates a new logger instance with the given configuration
func CreateLogger(config LoggerConfig) *slog.Logger {
	h := handler.NewCustomHandler(os.Stdout, config.Level, config.AddSource, config.SentryEnabled)
	return slog.New(h)
}
