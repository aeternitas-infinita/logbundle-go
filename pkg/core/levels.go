package core

import (
	"log/slog"
	"os"
	"strings"
)

// GetLvlFromEnv retrieves log level from environment variable
// Returns slog.LevelWarn as default if variable is not set or value is invalid
func GetLvlFromEnv(key string) slog.Level {
	if value := os.Getenv(key); value != "" {
		return GetLvlFromStr(value)
	}
	return slog.LevelWarn
}

// GetLvlFromStr converts string to slog.Level
// Accepts: "debug", "info", "warn", "error" (case-sensitive)
// Returns slog.LevelWarn for invalid values
func GetLvlFromStr(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}

// GetBoolFromStr converts string to boolean (case-insensitive)
// Returns true only if string is "true" (case-insensitive), false otherwise
func GetBoolFromStr(s string) bool {
	return strings.ToLower(s) == "true"
}
