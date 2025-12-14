package lgerr

// ErrorOption is a functional option for creating Error instances with flexible configuration
type ErrorOption func(*Error)

// WithMessage sets the error message
func WithMessage(message string) ErrorOption {
	return func(e *Error) {
		e.message = message
	}
}

// WithType sets the error type
func WithType(errType ErrorType) ErrorOption {
	return func(e *Error) {
		e.errorType = errType
	}
}

// WithHTTPStatusOpt sets the HTTP status code
func WithHTTPStatusOpt(status int) ErrorOption {
	return func(e *Error) {
		e.httpStatus = &status
	}
}

// WithTitle sets the public-facing title
func WithTitle(title string) ErrorOption {
	return func(e *Error) {
		e.title = title
	}
}

// WithDetail sets the public-facing detail
func WithDetail(detail string) ErrorOption {
	return func(e *Error) {
		e.detail = detail
	}
}

// WithContextKV adds a key-value pair to the error context
func WithContextKV(key string, value any) ErrorOption {
	return func(e *Error) {
		if e.context == nil {
			e.context = make(map[string]any)
		}
		e.context[key] = value
	}
}

// WithContextMap adds multiple context values at once
func WithContextMap(ctx map[string]any) ErrorOption {
	return func(e *Error) {
		if e.context == nil {
			e.context = make(map[string]any, len(ctx))
		}
		for k, v := range ctx {
			e.context[k] = v
		}
	}
}

// WithWrapped wraps another error
func WithWrapped(err error) ErrorOption {
	return func(e *Error) {
		e.wrapped = err
	}
}

// WithIgnoreSentry marks the error to skip Sentry reporting
func WithIgnoreSentry() ErrorOption {
	return func(e *Error) {
		e.ignoreSentry = true
	}
}

// WithValidationErr adds a validation error
func WithValidationErr(field, message string, value ...any) ErrorOption {
	return func(e *Error) {
		if e.validationErrors == nil {
			e.validationErrors = make([]ValidationError, 0, 4)
		}
		ve := ValidationError{
			Field:   field,
			Message: message,
		}
		if len(value) > 0 {
			ve.Value = value[0]
		}
		e.validationErrors = append(e.validationErrors, ve)
	}
}

// WithValidationErrs adds multiple validation errors
func WithValidationErrs(errors []ValidationError) ErrorOption {
	return func(e *Error) {
		e.validationErrors = errors
	}
}

// NewWithOptions creates an error with functional options pattern
// This provides a more flexible and extensible API for creating errors
//
// Example:
//
//	err := lgerr.NewWithOptions(
//	    lgerr.WithMessage("User not found"),
//	    lgerr.WithType(lgerr.TypeNotFound),
//	    lgerr.WithContextKV("user_id", userID),
//	    lgerr.WithContextKV("requested_by", requestorID),
//	    lgerr.WithTitle("User Not Found"),
//	    lgerr.WithDetail("The requested user does not exist"),
//	)
func NewWithOptions(opts ...ErrorOption) *Error {
	err := New("") // Creates base error with stack trace

	// Apply all options
	for _, opt := range opts {
		opt(err)
	}

	return err
}
