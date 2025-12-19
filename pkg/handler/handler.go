package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

// internalLog is used for logging within logbundle package (without source info for performance)
var internalLog = slog.New(NewCustomHandler(os.Stdout, slog.LevelError, false))

// CustomHandler implements slog.Handler with custom formatting
// Format: "YYYY/MM/DD HH:MM:SS [LEVEL] [file:line] message key=value..."
type CustomHandler struct {
	writer    io.Writer  // Output destination (typically os.Stdout)
	addSource bool       // Whether to include source file/line in output
	level     slog.Level // Minimum level to log
}

func NewCustomHandler(w io.Writer, level slog.Level, addSource bool) *CustomHandler {
	return &CustomHandler{
		writer:    w,
		level:     level,
		addSource: addSource,
	}
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle processes a log record and writes it to the output
// This is the core slog.Handler method
func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	const timestampFormat = "2006/01/02 15:04:05"
	timestamp := r.Time.Format(timestampFormat)
	level := fmt.Sprintf("[%s]", strings.ToUpper(r.Level.String()))

	var parts []string

	if h.addSource {
		var file string
		var line int

		// Check for manually provided source attribute
		var manualSource *slog.Source
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == "source" {
				if src, ok := a.Value.Any().(slog.Source); ok {
					manualSource = &src
					return false
				}
			}
			return true
		})

		if manualSource != nil {
			file = manualSource.File
			line = manualSource.Line
		} else if r.PC != 0 {
			frames := runtime.CallersFrames([]uintptr{r.PC})
			frame, _ := frames.Next()
			file = frame.File
			line = frame.Line
		}

		if file != "" {
			source := fmt.Sprintf("[%s:%d]", file, line)
			parts = append(parts, timestamp, level, source, r.Message)
		} else {
			parts = append(parts, timestamp, level, r.Message)
		}
	} else {
		parts = append(parts, timestamp, level, r.Message)
	}

	// Collect attributes in a single iteration
	attrs := make([]string, 0, 8) // Pre-allocate for typical attribute count
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "source" {
			return true // Skip source attribute as it's already handled
		}
		attrs = append(attrs, fmt.Sprintf("%s=%s", a.Key, a.Value.String()))
		return true
	})

	// Use strings.Builder for efficient concatenation
	var builder strings.Builder
	builder.WriteString(strings.Join(parts, " "))
	if len(attrs) > 0 {
		builder.WriteString(" ")
		builder.WriteString(strings.Join(attrs, " "))
	}

	_, err := fmt.Fprintln(h.writer, builder.String())
	return err
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Create a new handler with the same configuration
	// Note: This is a simplified implementation. For production use,
	// consider implementing proper attribute chaining if needed.
	return &CustomHandler{
		writer:    h.writer,
		level:     h.level,
		addSource: h.addSource,
	}
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	// Create a new handler with the same configuration
	// Note: This is a simplified implementation. For production use,
	// consider implementing proper group support if needed.
	return &CustomHandler{
		writer:    h.writer,
		level:     h.level,
		addSource: h.addSource,
	}
}

// GetInternalLogger returns the internal logger used by logbundle (without source)
func GetInternalLogger() *slog.Logger {
	return internalLog
}
