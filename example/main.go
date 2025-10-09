package main

import (
	"log/slog"

	"github.com/aeternitas-infinita/logbundle-go"
)

func main() {
	// Initialize logger with source info enabled
	logbundle.InitLog(logbundle.LoggerConfig{
		Level:         slog.LevelDebug,
		SentryEnabled: false,
		AddSource:     true,
	})

	// These logs should show the actual call location in this file
	logbundle.Debug("This is a debug message")
	logbundle.Info("This is an info message", slog.String("key", "value"))
	logbundle.Warn("This is a warning message")
	logbundle.Error("This is an error message", slog.Int("code", 500))

	// Call from another function
	testFunction()
}

func testFunction() {
	logbundle.Info("Called from testFunction", slog.String("location", "testFunction"))
}
