package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgsentry"
)

// Log is a package-level logger for internal use
var Log = slog.New(NewCustomHandler(os.Stdout, slog.LevelError, false, false))

// CustomHandler implements slog.Handler interface with Sentry integration
type CustomHandler struct {
	writer       io.Writer
	addSource    bool
	level        slog.Level
	enableSentry bool
}

// NewCustomHandler creates a new CustomHandler with the specified configuration
func NewCustomHandler(w io.Writer, level slog.Level, addSource, enableSentry bool) *CustomHandler {
	return &CustomHandler{
		writer:       w,
		level:        level,
		addSource:    addSource,
		enableSentry: enableSentry,
	}
}

// Enabled reports whether the handler handles records at the given level
func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle formats and outputs a log record, optionally sending it to Sentry
func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	const timestampFormat = "2006/01/02 15:04:05"
	timestamp := r.Time.Format(timestampFormat)

	level := fmt.Sprintf("[%s]", strings.ToUpper(r.Level.String()))

	var parts []string

	// Extract trace_id from context if available for log correlation
	var traceIDAttr *slog.Attr
	if ctx != nil {
		if traceID := core.GetLogTraceID(ctx); traceID != "" {
			attr := slog.String(core.TraceIDKey, traceID)
			traceIDAttr = &attr
		}
	}

	if h.addSource {
		var file string
		var line int

		// First check if there's a manually set source in the attributes
		var manualSource *slog.Source
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == "source" {
				if src, ok := a.Value.Any().(slog.Source); ok {
					manualSource = &src
					return false // stop iteration
				}
			}
			return true
		})

		if manualSource != nil {
			file = manualSource.File
			line = manualSource.Line
		} else if r.PC != 0 {
			// r.PC is set by slog when AddSource is enabled
			// It points to the actual caller, not the library code
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

	var attrs []string

	// Add trace_id first if available
	if traceIDAttr != nil {
		attrs = append(attrs, fmt.Sprintf("%s=%s", traceIDAttr.Key, traceIDAttr.Value.String()))
	}

	r.Attrs(func(a slog.Attr) bool {
		// Skip the manual source attribute as it's already handled
		if a.Key == "source" {
			return true
		}
		attrs = append(attrs, fmt.Sprintf("%s=%s", a.Key, a.Value.String()))
		return true
	})

	var slogAttrs []slog.Attr
	r.Attrs(func(attr slog.Attr) bool {
		slogAttrs = append(slogAttrs, attr)
		return true
	})

	logLine := strings.Join(parts, " ")
	if len(attrs) > 0 {
		logLine += " " + strings.Join(attrs, " ")
	}

	_, err := fmt.Fprintln(h.writer, logLine)
	if err != nil {
		return err
	}

	// Send to Sentry if enabled
	if h.enableSentry {
		lgsentry.CaptureEventForSlog(ctx, r, slogAttrs)
	}

	return nil
}

// WithAttrs returns a new handler with additional attributes
// Currently returns the same handler (attributes not persisted)
func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns a new handler with a group name
// Currently returns the same handler (groups not supported)
func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return h
}
