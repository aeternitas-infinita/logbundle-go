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
	err := New(message)
	err.errorType = TypeValidation
	err.title = "Validation Error"

	for _, opt := range opts {
		opt(err)
	}
	return err
}

// Database creates a database error
func Database(message string, opts ...ErrorOption) *Error {
	err := New(message)
	err.errorType = TypeDatabase
	err.title = "Database Error"

	for _, opt := range opts {
		opt(err)
	}
	return err
}

// Internal creates an internal server error
func Internal(message string, opts ...ErrorOption) *Error {
	err := New(message)
	err.errorType = TypeInternal
	err.title = "Internal Server Error"

	for _, opt := range opts {
		opt(err)
	}
	return err
}

// Forbidden creates a forbidden access error
func Forbidden(resource string, reason string, opts ...ErrorOption) *Error {
	err := New(fmt.Sprintf("access forbidden: %s", reason))
	err.errorType = TypeForbidden
	err.title = "Access Forbidden"
	err.detail = reason
	if err.context == nil {
		err.context = make(map[string]any, 2)
	}
	err.context["resource"] = resource
	err.context["reason"] = reason

	for _, opt := range opts {
		opt(err)
	}
	return err
}

// Unauthorized creates an unauthorized error
func Unauthorized(reason string, opts ...ErrorOption) *Error {
	err := New(fmt.Sprintf("unauthorized: %s", reason))
	err.errorType = TypeUnauth
	err.title = "Unauthorized"
	err.detail = reason
	if err.context == nil {
		err.context = make(map[string]any, 1)
	}
	err.context["reason"] = reason

	for _, opt := range opts {
		opt(err)
	}
	return err
}

// BadInput creates a bad input error
func BadInput(message string, opts ...ErrorOption) *Error {
	err := New(message)
	err.errorType = TypeBadInput
	err.title = "Bad Request"

	for _, opt := range opts {
		opt(err)
	}
	return err
}

// Conflict creates a resource conflict error
func Conflict(resource string, reason string, opts ...ErrorOption) *Error {
	err := New(fmt.Sprintf("%s conflict: %s", resource, reason))
	err.errorType = TypeConflict
	err.title = "Resource Conflict"
	err.detail = reason
	if err.context == nil {
		err.context = make(map[string]any, 2)
	}
	err.context["resource"] = resource
	err.context["reason"] = reason

	for _, opt := range opts {
		opt(err)
	}
	return err
}

// External creates an external service error
func External(service string, message string, opts ...ErrorOption) *Error {
	err := New(fmt.Sprintf("external service error: %s - %s", service, message))
	err.errorType = TypeExternal
	err.title = "External Service Error"
	err.detail = message
	if err.context == nil {
		err.context = make(map[string]any, 1)
	}
	err.context["service"] = service

	for _, opt := range opts {
		opt(err)
	}
	return err
}

// Timeout creates a timeout error
func Timeout(operation string, duration string, opts ...ErrorOption) *Error {
	err := New(fmt.Sprintf("timeout: %s exceeded %s", operation, duration))
	err.errorType = TypeTimeout
	err.title = "Request Timeout"
	err.detail = fmt.Sprintf("Operation %s exceeded timeout of %s", operation, duration)
	if err.context == nil {
		err.context = make(map[string]any, 2)
	}
	err.context["operation"] = operation
	err.context["duration"] = duration

	for _, opt := range opts {
		opt(err)
	}
	return err
}

// Busy creates a service busy/unavailable error
func Busy(message string, opts ...ErrorOption) *Error {
	err := New(message)
	err.errorType = TypeBusy
	err.title = "Service Unavailable"

	for _, opt := range opts {
		opt(err)
	}
	return err
}
