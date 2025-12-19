package lgerr

type ErrorOption func(*Error)

func WithMessage(message string) ErrorOption {
	return func(e *Error) {
		e.message = message
	}
}

func WithType(errType ErrorType) ErrorOption {
	return func(e *Error) {
		e.errorType = errType
	}
}

func WithHTTPStatusOpt(status int) ErrorOption {
	return func(e *Error) {
		e.httpStatus = &status
	}
}

func WithTitle(title string) ErrorOption {
	return func(e *Error) {
		e.title = title
	}
}

func WithDetail(detail string) ErrorOption {
	return func(e *Error) {
		e.detail = detail
	}
}

func WithContext(key string, value any) ErrorOption {
	return func(e *Error) {
		if e.context == nil {
			e.context = make(map[string]any)
		}
		e.context[key] = value
	}
}

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

func WithWrapped(err error) ErrorOption {
	return func(e *Error) {
		e.wrapped = err
	}
}

func WithIgnoreSentry() ErrorOption {
	return func(e *Error) {
		e.ignoreSentry = true
	}
}

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

func WithValidationErrs(errors []ValidationError) ErrorOption {
	return func(e *Error) {
		e.validationErrors = errors
	}
}

func NewWithOptions(opts ...ErrorOption) *Error {
	err := New("")

	for _, opt := range opts {
		opt(err)
	}

	return err
}
