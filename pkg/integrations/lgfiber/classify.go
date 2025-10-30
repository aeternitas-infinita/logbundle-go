package lgfiber

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/erri"
)

// Common error patterns for classification
const (
	errPatternConnection  = "connection"
	errPatternTimeout     = "timeout"
	errPatternNotFound    = "not found"
	errPatternUnauth      = "unauthorized"
	errPatternForbidden   = "forbidden"
	errPatternValidation  = "validation"
	errPatternDatabase    = "database"
	maxFingerprintLen     = 50
)

// getErrorType classifies an error into a category for monitoring/alerting
func getErrorType(err error) string {
	// Check custom internal error types first
	var internalErr *erri.Erri
	if errors.As(err, &internalErr) {
		return mapErriTypeToString(internalErr.Type)
	}

	// Check Fiber errors
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return "fiber_error"
	}

	// Check standard context errors
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "canceled"
	}

	// Fallback to pattern matching on error message
	return classifyErrorByMessage(err.Error())
}

// getErrorFingerprint generates a consistent fingerprint for error grouping in Sentry
func getErrorFingerprint(err error) string {
	// Check custom internal error types first
	var internalErr *erri.Erri
	if errors.As(err, &internalErr) {
		if internalErr.Property != "" {
			return fmt.Sprintf("%s-%s", string(internalErr.Type), internalErr.Property)
		}
		return string(internalErr.Type)
	}

	// Check Fiber errors
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fmt.Sprintf("fiber-%d", fiberErr.Code)
	}

	// Pattern-based fingerprinting
	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "connection refused"):
		return "connection-refused"
	case strings.Contains(errStr, errPatternTimeout):
		return errPatternTimeout
	case strings.Contains(errStr, errPatternNotFound):
		return "not-found"
	case strings.Contains(errStr, errPatternUnauth):
		return "unauthorized"
	case strings.Contains(errStr, errPatternForbidden):
		return errPatternForbidden
	default:
		// Truncate long error messages for fingerprinting
		if len(errStr) > maxFingerprintLen {
			return errStr[:maxFingerprintLen]
		}
		return errStr
	}
}

// mapErriTypeToString converts erri error types to string representations
func mapErriTypeToString(typ erri.ErriType) string {
	switch typ {
	case erri.ErriStruct.NOT_FOUND:
		return "not_found"
	case erri.ErriStruct.VALIDATION:
		return "validation"
	case erri.ErriStruct.DATABASE:
		return "database"
	case erri.ErriStruct.INTERNAL:
		return "internal"
	case erri.ErriStruct.BUSY:
		return "busy"
	case erri.ErriStruct.FORBIDDEN:
		return "forbidden"
	case erri.ErriStruct.WRONG_INPUT:
		return "wrong_input"
	default:
		return "internal_error_unknown"
	}
}

// classifyErrorByMessage attempts to classify an error based on its message
func classifyErrorByMessage(errMsg string) string {
	errLower := strings.ToLower(errMsg)
	switch {
	case strings.Contains(errLower, errPatternConnection):
		return "connection"
	case strings.Contains(errLower, errPatternTimeout):
		return "timeout"
	case strings.Contains(errLower, errPatternNotFound):
		return "not_found"
	case strings.Contains(errLower, errPatternUnauth):
		return "unauthorized"
	case strings.Contains(errLower, errPatternForbidden):
		return "forbidden"
	case strings.Contains(errLower, errPatternValidation):
		return "validation"
	case strings.Contains(errLower, errPatternDatabase):
		return "database"
	default:
		return "unknown"
	}
}
