package core

import (
	"log/slog"
)

var TraceIDKey = "trace_id"

type LoggerConfig struct {
	AddSource    bool
	Level        slog.Level
	EnableSentry bool
}

type CoreLogger struct {
	logger *slog.Logger
	config *LoggerConfig
}

func NewCoreLogger(handler slog.Handler, config *LoggerConfig) *CoreLogger {
	return &CoreLogger{
		logger: slog.New(handler),
		config: config,
	}
}
