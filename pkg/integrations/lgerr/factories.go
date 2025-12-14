package lgerr

import "fmt"

// NotFound creates a "not found" error with resource context
func NotFound(resource string, id any) *Error {
	return NewWithOptions(
		WithMessage(fmt.Sprintf("%s not found", resource)),
		WithType(TypeNotFound),
		WithContextKV("resource", resource),
		WithContextKV("resource_id", id),
		WithTitle("Resource Not Found"),
		WithDetail(fmt.Sprintf("The requested %s does not exist", resource)),
	)
}

// Validation creates a validation error
func Validation(message string, opts ...ErrorOption) *Error {
	options := []ErrorOption{
		WithMessage(message),
		WithType(TypeValidation),
		WithTitle("Validation Error"),
	}
	return NewWithOptions(append(options, opts...)...)
}

// Database creates a database error
func Database(message string, opts ...ErrorOption) *Error {
	options := []ErrorOption{
		WithMessage(message),
		WithType(TypeDatabase),
		WithTitle("Database Error"),
	}
	return NewWithOptions(append(options, opts...)...)
}

// Internal creates an internal server error
func Internal(message string, opts ...ErrorOption) *Error {
	options := []ErrorOption{
		WithMessage(message),
		WithType(TypeInternal),
		WithTitle("Internal Server Error"),
	}
	return NewWithOptions(append(options, opts...)...)
}

// Forbidden creates a forbidden access error
func Forbidden(resource string, reason string, opts ...ErrorOption) *Error {
	options := []ErrorOption{
		WithMessage(fmt.Sprintf("access forbidden: %s", reason)),
		WithType(TypeForbidden),
		WithContextKV("resource", resource),
		WithContextKV("reason", reason),
		WithTitle("Access Forbidden"),
		WithDetail(reason),
	}
	return NewWithOptions(append(options, opts...)...)
}

// Unauthorized creates an unauthorized error
func Unauthorized(reason string, opts ...ErrorOption) *Error {
	options := []ErrorOption{
		WithMessage(fmt.Sprintf("unauthorized: %s", reason)),
		WithType(TypeUnauth),
		WithContextKV("reason", reason),
		WithTitle("Unauthorized"),
		WithDetail(reason),
	}
	return NewWithOptions(append(options, opts...)...)
}

// BadInput creates a bad input error
func BadInput(message string, opts ...ErrorOption) *Error {
	options := []ErrorOption{
		WithMessage(message),
		WithType(TypeBadInput),
		WithTitle("Bad Request"),
	}
	return NewWithOptions(append(options, opts...)...)
}

// Conflict creates a resource conflict error
func Conflict(resource string, reason string, opts ...ErrorOption) *Error {
	options := []ErrorOption{
		WithMessage(fmt.Sprintf("%s conflict: %s", resource, reason)),
		WithType(TypeConflict),
		WithContextKV("resource", resource),
		WithContextKV("reason", reason),
		WithTitle("Resource Conflict"),
		WithDetail(reason),
	}
	return NewWithOptions(append(options, opts...)...)
}

// External creates an external service error
func External(service string, message string, opts ...ErrorOption) *Error {
	options := []ErrorOption{
		WithMessage(fmt.Sprintf("external service error: %s - %s", service, message)),
		WithType(TypeExternal),
		WithContextKV("service", service),
		WithTitle("External Service Error"),
		WithDetail(message),
	}
	return NewWithOptions(append(options, opts...)...)
}

// Timeout creates a timeout error
func Timeout(operation string, duration string, opts ...ErrorOption) *Error {
	options := []ErrorOption{
		WithMessage(fmt.Sprintf("timeout: %s exceeded %s", operation, duration)),
		WithType(TypeTimeout),
		WithContextKV("operation", operation),
		WithContextKV("duration", duration),
		WithTitle("Request Timeout"),
		WithDetail(fmt.Sprintf("Operation %s exceeded timeout of %s", operation, duration)),
	}
	return NewWithOptions(append(options, opts...)...)
}

// Busy creates a service busy/unavailable error
func Busy(message string, opts ...ErrorOption) *Error {
	options := []ErrorOption{
		WithMessage(message),
		WithType(TypeBusy),
		WithTitle("Service Unavailable"),
	}
	return NewWithOptions(append(options, opts...)...)
}
