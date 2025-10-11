package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgsentry"
)

var Log = slog.New(NewCustomHandler(os.Stdout, slog.LevelError, false, false))

type CustomHandler struct {
	writer       io.Writer
	addSource    bool
	level        slog.Level
	enableSentry bool
}

func NewCustomHandler(w io.Writer, level slog.Level, addSource, enableSentry bool) *CustomHandler {
	return &CustomHandler{
		writer:       w,
		level:        level,
		addSource:    addSource,
		enableSentry: enableSentry,
	}
}

func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	timestamp := r.Time.Format("2006/01/02 15:04:05")

	level := fmt.Sprintf("[%s]", strings.ToUpper(r.Level.String()))

	var parts []string

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

	if h.enableSentry == true {
		lgsentry.CaptureEvent(ctx, r, slogAttrs)
	}

	return nil
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return h
}
