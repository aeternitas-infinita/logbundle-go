package lgerr

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

type ErrorType string

const (
	TypeInternal   ErrorType = "internal"
	TypeNotFound   ErrorType = "not_found"
	TypeValidation ErrorType = "validation"
	TypeDatabase   ErrorType = "database"
	TypeBusy       ErrorType = "busy"
	TypeForbidden  ErrorType = "forbidden"
	TypeBadInput   ErrorType = "bad_input"
	TypeUnauth     ErrorType = "unauthorized"
	TypeConflict   ErrorType = "conflict"
	TypeExternal   ErrorType = "external"
	TypeTimeout    ErrorType = "timeout"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

type ErrorResponse struct {
	Title  string            `json:"title"`
	Detail string            `json:"detail,omitempty"`
	Errors []ValidationError `json:"errors,omitempty"`
	Meta   map[string]any    `json:"meta,omitempty"`
}

type Response[T any] struct {
	Data T `json:"data,omitempty"`
}

type Error struct {
	message          string
	title            string
	detail           string
	errorType        ErrorType
	httpStatus       *int
	context          map[string]any
	file             string
	line             int
	stackTrace       []uintptr
	wrapped          error
	ignoreSentry     bool
	validationErrors []ValidationError
}

var (
	httpStatusMap     map[ErrorType]int
	customTypeMapping map[ErrorType]int
	mapMutex          sync.RWMutex
)

func init() {
	httpStatusMap = map[ErrorType]int{
		TypeInternal:   500,
		TypeNotFound:   404,
		TypeValidation: 400,
		TypeDatabase:   500,
		TypeBusy:       503,
		TypeForbidden:  403,
		TypeBadInput:   400,
		TypeUnauth:     401,
		TypeConflict:   409,
		TypeExternal:   502,
		TypeTimeout:    504,
	}
}

func RegisterErrorType(errType ErrorType, httpStatus int) {
	mapMutex.Lock()
	defer mapMutex.Unlock()

	if customTypeMapping == nil {
		customTypeMapping = make(map[ErrorType]int)
	}

	customTypeMapping[errType] = httpStatus
}

func SetHTTPStatusMap(customMap map[ErrorType]int) {
	mapMutex.Lock()
	defer mapMutex.Unlock()

	if customTypeMapping == nil {
		customTypeMapping = make(map[ErrorType]int)
	}

	for errType, status := range customMap {
		customTypeMapping[errType] = status
	}
}

func ResetHTTPStatusMap() {
	mapMutex.Lock()
	defer mapMutex.Unlock()

	customTypeMapping = make(map[ErrorType]int)
}

func GetHTTPStatus(errType ErrorType) int {
	return getHTTPStatus(errType)
}

func getHTTPStatus(errType ErrorType) int {
	mapMutex.RLock()
	defer mapMutex.RUnlock()

	if customTypeMapping != nil {
		if status, ok := customTypeMapping[errType]; ok {
			return status
		}
	}

	if status, ok := httpStatusMap[errType]; ok {
		return status
	}

	return 500
}

func New(message string) *Error {
	const maxStackDepth = 32
	var pcs [maxStackDepth]uintptr
	n := runtime.Callers(2, pcs[:])

	file := "unknown"
	line := 0

	if n > 0 {
		frames := runtime.CallersFrames(pcs[:n])
		if frame, more := frames.Next(); more || frame.PC != 0 {
			file = frame.File
			line = frame.Line
		}
	}

	return &Error{
		message:    message,
		errorType:  TypeInternal,
		file:       file,
		line:       line,
		stackTrace: pcs[:n:n],
	}
}

func (e *Error) WithType(errType ErrorType) *Error {
	e.errorType = errType
	return e
}

func (e *Error) WithContext(key string, value any) *Error {
	if e.context == nil {
		e.context = make(map[string]any)
	}
	e.context[key] = value
	return e
}

func (e *Error) WithHTTPStatus(status int) *Error {
	e.httpStatus = &status
	return e
}

func (e *Error) Wrap(err error) *Error {
	e.wrapped = err
	return e
}

func (e *Error) SetHTTPStatus(status int) {
	e.httpStatus = &status
}

func (e *Error) IgnoreSentry() *Error {
	e.ignoreSentry = true
	return e
}

func (e *Error) ShouldIgnoreSentry() bool {
	return e.ignoreSentry
}

func (e *Error) WithTitle(title string) *Error {
	e.title = title
	return e
}

func (e *Error) WithDetail(detail string) *Error {
	e.detail = detail
	return e
}

func (e *Error) WithValidationError(field string, message string, value ...any) *Error {
	if e.validationErrors == nil {
		e.validationErrors = make([]ValidationError, 0, 4) // Pre-allocate for typical validation error count
	}
	validationErr := ValidationError{
		Field:   field,
		Message: message,
	}
	if len(value) > 0 {
		validationErr.Value = value[0]
	}
	e.validationErrors = append(e.validationErrors, validationErr)
	return e
}

func (e *Error) WithValidationErrors(errors []ValidationError) *Error {
	e.validationErrors = errors
	return e
}

func (e *Error) Error() string {
	if e.wrapped != nil {
		return fmt.Sprintf("%s: %v", e.message, e.wrapped)
	}
	return e.message
}

func (e *Error) Unwrap() error {
	return e.wrapped
}

func (e *Error) Message() string {
	return e.message
}

func (e *Error) Type() ErrorType {
	return e.errorType
}

func (e *Error) HTTPStatus() int {
	if e.httpStatus != nil {
		return *e.httpStatus
	}
	return getHTTPStatus(e.errorType)
}

func (e *Error) Context() map[string]any {
	return e.context
}

func (e *Error) File() string {
	return e.file
}

func (e *Error) Line() int {
	return e.line
}

func (e *Error) Wrapped() error {
	return e.wrapped
}

func (e *Error) Title() string {
	return e.title
}

func (e *Error) Detail() string {
	return e.detail
}

func (e *Error) ValidationErrors() []ValidationError {
	return e.validationErrors
}

func (e *Error) HasValidationErrors() bool {
	return len(e.validationErrors) > 0
}

func (e *Error) ToErrorResponse() ErrorResponse {
	response := ErrorResponse{
		Title:  e.title,
		Detail: e.detail,
		Errors: e.validationErrors,
	}

	if len(e.context) > 0 {
		response.Meta = e.context
	}

	return response
}

func (e *Error) StackTrace() []uintptr {
	return e.stackTrace
}

func (e *Error) StackFrames() *runtime.Frames {
	return runtime.CallersFrames(e.stackTrace)
}

func (e *Error) FormatStackTrace() string {
	if len(e.stackTrace) == 0 {
		return "no stack trace available"
	}

	var builder strings.Builder
	// Pre-allocate approximate size: ~100 chars per frame
	builder.Grow(len(e.stackTrace) * 100)

	frames := runtime.CallersFrames(e.stackTrace)
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&builder, "%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return builder.String()
}
