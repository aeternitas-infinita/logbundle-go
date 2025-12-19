package config

import (
	"log/slog"
	"sync"
)

var (
	middlewareLogger      *slog.Logger
	middlewareLoggerMutex sync.RWMutex
)

// SetMiddlewareLogger sets the logger to be used by all middlewares
// If not set, middlewares will use the internal logger
func SetMiddlewareLogger(logger *slog.Logger) {
	middlewareLoggerMutex.Lock()
	middlewareLogger = logger
	middlewareLoggerMutex.Unlock()
}

// GetMiddlewareLogger returns the configured middleware logger, or nil if not set
func GetMiddlewareLogger() *slog.Logger {
	middlewareLoggerMutex.RLock()
	defer middlewareLoggerMutex.RUnlock()
	return middlewareLogger
}
