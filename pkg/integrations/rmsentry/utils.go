package rmsentry

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

type SourceInfo struct {
	File string
	Line int
}

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

func extractSentryData(attrs []slog.Attr) (map[string]string, map[string]interface{}, error) {
	tags := make(map[string]string)
	extra := make(map[string]interface{})
	var errorValue error

	for _, atr := range attrs {
		key := atr.Key
		value := atr.Value.Any()

		if err, ok := value.(error); ok && errorValue == nil {
			errorValue = err
			continue
		}

		if strVal, ok := value.(string); ok && len(strVal) < 100 && !strings.Contains(strVal, "\n") {
			tags[key] = strVal
		} else if numVal, ok := value.(int); ok {
			tags[key] = fmt.Sprintf("%d", numVal)
		} else if numVal, ok := value.(int64); ok {
			tags[key] = fmt.Sprintf("%d", numVal)
		} else if boolVal, ok := value.(bool); ok {
			tags[key] = fmt.Sprintf("%t", boolVal)
		} else {
			extra[key] = value
		}

	}

	return tags, extra, errorValue
}
