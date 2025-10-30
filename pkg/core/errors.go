package core

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

// ErrAttr creates a slog.Attr for an error with key "error"
func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

// GetLinePositionStringWithSkip returns formatted file:line position string
// skip indicates how many stack frames to skip (caller depth)
func GetLinePositionStringWithSkip(skip int) string {
	_, file, line, _ := runtime.Caller(skip)
	return fmt.Sprintf("[%s:%d]", file, line)
}

// ExtractErrorLocationWithDetails parses a stack trace and returns the first non-library error location
// Returns: (fullLocation, file, lineNumber)
func ExtractErrorLocationWithDetails(stackTrace string) (string, string, int) {
	lines := strings.Split(stackTrace, "\n")

	// Skip goroutine line and runtime panic lines
	// Looking for the first non-runtime, non-library line
	for i := 0; i < len(lines)-1; i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip goroutine info
		if strings.HasPrefix(line, "goroutine ") {
			continue
		}

		// Look at the next line which should contain file:line
		if i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])

			// Check if this line contains .go: (file path with line number)
			if strings.Contains(nextLine, ".go:") {
				// Normalize path for checking (handle both Windows and Unix paths)
				normalizedPath := strings.ReplaceAll(nextLine, "\\", "/")

				// Skip runtime and library internals
				if strings.Contains(normalizedPath, "runtime/") ||
					strings.Contains(normalizedPath, "/runtime.") ||
					strings.Contains(normalizedPath, "logbundle-go/") ||
					strings.Contains(normalizedPath, "/logbundle-go/") ||
					strings.Contains(normalizedPath, "\\logbundle-go\\") ||
					strings.Contains(line, "FiberRecoverMiddleware") ||
					strings.Contains(line, "RecoverMiddleware") ||
					strings.Contains(line, "RecoverWithContext") ||
					strings.Contains(line, "panic") ||
					strings.Contains(line, "(*Ctx).Next") {
					continue
				}

				// Extract file path
				parts := strings.Fields(nextLine)
				if len(parts) > 0 {
					filePath := parts[0]

					// Parse file and line number
					fileLineParts := strings.Split(filePath, ":")
					if len(fileLineParts) >= 2 {
						file := fileLineParts[0]
						lineNum := 0
						fmt.Sscanf(fileLineParts[1], "%d", &lineNum)
						return filePath, file, lineNum
					}

					return filePath, filePath, 0
				}
			}
		}
	}

	return "unknown location", "", 0
}
