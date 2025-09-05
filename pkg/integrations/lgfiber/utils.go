package lgfiber

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/erri"
)

func getErrorType(err error) string {
	var internalErr *erri.Erri
	if errors.As(err, &internalErr) {
		switch internalErr.Type {
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
		default:
			return "internal_error_unknown"
		}
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return "fiber_error"
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "canceled"
	default:
		errStr := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errStr, "connection"):
			return "connection"
		case strings.Contains(errStr, "timeout"):
			return "timeout"
		case strings.Contains(errStr, "not found"):
			return "not_found"
		case strings.Contains(errStr, "unauthorized"):
			return "unauthorized"
		case strings.Contains(errStr, "forbidden"):
			return "forbidden"
		case strings.Contains(errStr, "validation"):
			return "validation"
		case strings.Contains(errStr, "database"):
			return "database"
		default:
			return "unknown"
		}
	}
}

func getErrorFingerprint(err error) string {
	var internalErr *erri.Erri
	if errors.As(err, &internalErr) {
		if internalErr.Property != "" {
			return fmt.Sprintf("%s-%s", string(internalErr.Type), internalErr.Property)
		}
		return string(internalErr.Type)
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fmt.Sprintf("fiber-%d", fiberErr.Code)
	}

	errStr := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errStr, "connection refused"):
		return "connection-refused"
	case strings.Contains(errStr, "timeout"):
		return "timeout"
	case strings.Contains(errStr, "not found"):
		return "not-found"
	case strings.Contains(errStr, "unauthorized"):
		return "unauthorized"
	case strings.Contains(errStr, "forbidden"):
		return "forbidden"
	default:
		if len(errStr) > 50 {
			return errStr[:50]
		}
		return errStr
	}
}
