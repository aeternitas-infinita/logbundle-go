package core

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}

func GetLinePositionStringWithSkip(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "[unknown:0]"
	}
	return fmt.Sprintf("[%s:%d]", file, line)
}

// shouldSkipFrame determines if a stack frame should be filtered out
func shouldSkipFrame(line, normalizedPath string) bool {
	// Skip runtime and internal frames
	internalPaths := []string{
		"runtime/",
		"/runtime.",
		"logbundle-go/",
		"/logbundle-go/",
		"\\logbundle-go\\",
	}

	for _, path := range internalPaths {
		if strings.Contains(normalizedPath, path) {
			return true
		}
	}

	// Skip middleware and panic frames
	skipFunctions := []string{
		"FiberRecoverMiddleware",
		"RecoverMiddleware",
		"RecoverWithContext",
		"panic",
		"(*Ctx).Next",
	}

	for _, fn := range skipFunctions {
		if strings.Contains(line, fn) {
			return true
		}
	}

	return false
}

// parseFileLocation extracts file path and line number from a stack trace line
func parseFileLocation(nextLine string) (filePath, file string, lineNum int) {
	parts := strings.Fields(nextLine)
	if len(parts) == 0 {
		return "", "", 0
	}

	filePath = parts[0]
	fileLineParts := strings.Split(filePath, ":")

	if len(fileLineParts) >= 2 {
		file = fileLineParts[0]
		fmt.Sscanf(fileLineParts[1], "%d", &lineNum)
		return filePath, file, lineNum
	}

	return filePath, filePath, 0
}

// ExtractErrorLocationWithDetails extracts the error location from a stack trace string,
// filtering out internal runtime and middleware frames to find the actual application code location
func ExtractErrorLocationWithDetails(stackTrace string) (string, string, int) {
	lines := strings.Split(stackTrace, "\n")

	for i := 0; i < len(lines)-1; i++ {
		line := strings.TrimSpace(lines[i])

		// Skip empty lines and goroutine headers
		if line == "" || strings.HasPrefix(line, "goroutine ") {
			continue
		}

		// Check if next line contains file location
		if i+1 < len(lines) {
			nextLine := strings.TrimSpace(lines[i+1])

			if strings.Contains(nextLine, ".go:") {
				normalizedPath := strings.ReplaceAll(nextLine, "\\", "/")

				// Skip internal and middleware frames
				if shouldSkipFrame(line, normalizedPath) {
					continue
				}

				return parseFileLocation(nextLine)
			}
		}
	}

	return "unknown location", "", 0
}
