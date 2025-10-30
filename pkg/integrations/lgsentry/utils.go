package lgsentry

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

const (
	// Maximum string length before treating as "extra" data instead of a tag
	maxTagLength = 100
)

// SourceInfo contains file and line information from a log record
type SourceInfo struct {
	File string
	Line int
}

// extractSourceInfo retrieves source file and line number from a slog record
func extractSourceInfo(r slog.Record) *SourceInfo {
	if r.PC == 0 {
		return nil
	}

	frames := runtime.CallersFrames([]uintptr{r.PC})
	frame, _ := frames.Next()

	if frame.File == "" {
		return nil
	}

	return &SourceInfo{
		File: frame.File,
		Line: frame.Line,
	}
}

// extractSentryData separates slog attributes into Sentry tags (indexed strings),
// extra data (complex objects), and extracts the first error value
func extractSentryData(attrs []slog.Attr) (map[string]string, map[string]any, error) {
	tags := make(map[string]string)
	extra := make(map[string]any)
	var errorValue error

	for _, atr := range attrs {
		key := atr.Key
		value := atr.Value.Any()

		// Extract first error encountered
		if err, ok := value.(error); ok && errorValue == nil {
			errorValue = err
			continue
		}

		// Short strings become tags (searchable in Sentry)
		if strVal, ok := value.(string); ok && len(strVal) < maxTagLength && !strings.Contains(strVal, "\n") {
			tags[key] = strVal
			continue
		}

		// Numeric values become tags
		switch numVal := value.(type) {
		case int:
			tags[key] = fmt.Sprintf("%d", numVal)
		case int64:
			tags[key] = fmt.Sprintf("%d", numVal)
		case bool:
			tags[key] = fmt.Sprintf("%t", numVal)
		default:
			// Everything else becomes extra data
			extra[key] = value
		}
	}

	return tags, extra, errorValue
}
