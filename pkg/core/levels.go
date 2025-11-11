package core

import (
	"log/slog"
	"os"
	"strings"
)

func GetLvlFromEnv(key string) slog.Level {
	if value := os.Getenv(key); value != "" {
		return GetLvlFromStr(value)
	}
	return slog.LevelWarn
}

func GetLvlFromStr(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelWarn
	}
}

func GetBoolFromStr(s string) bool {
	return strings.ToLower(s) == "true"
}
