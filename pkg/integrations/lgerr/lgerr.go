package lgerr

import (
	"fmt"
	"runtime"
	"sync"
)

// ErrorType represents the category of an error
type ErrorType string

const (
	TypeInternal   ErrorType = "internal"     // Internal server errors (500)
	TypeNotFound   ErrorType = "not_found"    // Resource not found (404)
	TypeValidation ErrorType = "validation"   // Validation failures (400)
	TypeDatabase   ErrorType = "database"     // Database errors (500)
	TypeBusy       ErrorType = "busy"         // Service busy/unavailable (503)
	TypeForbidden  ErrorType = "forbidden"    // Access forbidden (403)
	TypeBadInput   ErrorType = "bad_input"    // Bad request input (400)
	TypeUnauth     ErrorType = "unauthorized" // Unauthorized access (401)
	TypeConflict   ErrorType = "conflict"     // Resource conflict (409)
	TypeExternal   ErrorType = "external"     // External service error (502)
	TypeTimeout    ErrorType = "timeout"      // Request timeout (504)
)

// ValidationError represents a single field validation error
// Following RFC 7807 problem details standard
type ValidationError struct {
	Field   string `json:"field"`           // Field name that failed validation
	Message string `json:"message"`         // Human-readable error message
	Value   any    `json:"value,omitempty"` // The value that failed validation
}

// ErrorResponse represents the standard error response format
// Following RFC 7807 Problem Details for HTTP APIs
type ErrorResponse struct {
	Title  string                 `json:"title"`            // Human-readable summary of the error
	Detail string                 `json:"detail,omitempty"` // Human-readable explanation specific to this occurrence
	Errors []ValidationError      `json:"errors,omitempty"` // Validation errors for fields
	Meta   map[string]interface{} `json:"meta,omitempty"`   // Additional context/metadata
}

// Response is a generic response wrapper for successful API responses
type Response[T any] struct {
	Data T `json:"data,omitempty"` // The response payload
}

// Error represents a rich error with type, context, stack trace, and HTTP mapping
// Thread-safe for reading after creation. Use builder methods for initialization.
type Error struct {
	message          string            // Internal error message for logging
	title            string            // Public-facing title (RFC 7807)
	detail           string            // Public-facing detail (RFC 7807)
	errorType        ErrorType         // Category of error (affects HTTP status)
	httpStatus       *int              // Optional override for HTTP status code
	context          map[string]any    // Additional context metadata
	file             string            // Source file where error was created
	line             int               // Line number where error was created
	stackTrace       []uintptr         // Runtime stack trace
	wrapped          error             // Wrapped underlying error
	ignoreSentry     bool              // Flag to skip Sentry reporting
	validationErrors []ValidationError // Field-level validation errors
}

var (
	// httpStatusMap provides default HTTP status codes for each error type
	httpStatusMap map[ErrorType]int
	// customTypeMapping stores user-defined overrides for HTTP status codes
	customTypeMapping map[ErrorType]int
	// mapMutex protects concurrent access to HTTP status mappings
	mapMutex sync.RWMutex
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

// RegisterErrorType registers a custom error type with its HTTP status code
// This allows you to define your own error types beyond the built-in ones
//
// Example:
//
//	const TypeRateLimited ErrorType = "rate_limited"
//	lgerr.RegisterErrorType(TypeRateLimited, 429)
func RegisterErrorType(errType ErrorType, httpStatus int) {
	mapMutex.Lock()
	defer mapMutex.Unlock()

	if customTypeMapping == nil {
		customTypeMapping = make(map[ErrorType]int)
	}

	customTypeMapping[errType] = httpStatus
}

// SetHTTPStatusMap overrides HTTP status codes for existing or custom error types
// Use this to customize the HTTP status mapping for your application
//
// Example:
//
//	lgerr.SetHTTPStatusMap(map[lgerr.ErrorType]int{
//	    lgerr.TypeNotFound: 410,  // Use 410 Gone instead of 404
//	    lgerr.TypeBusy:     429,  // Use 429 Too Many Requests instead of 503
//	})
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

// ResetHTTPStatusMap resets all custom HTTP status mappings to defaults
func ResetHTTPStatusMap() {
	mapMutex.Lock()
	defer mapMutex.Unlock()

	customTypeMapping = make(map[ErrorType]int)
}

// GetHTTPStatus returns the HTTP status code for a given error type
// Useful for testing or checking current mappings
func GetHTTPStatus(errType ErrorType) int {
	return getHTTPStatus(errType)
}

func getHTTPStatus(errType ErrorType) int {
	mapMutex.RLock()
	defer mapMutex.RUnlock()

	// Check custom mapping first
	if customTypeMapping != nil {
		if status, ok := customTypeMapping[errType]; ok {
			return status
		}
	}

	// Then check default mapping
	if status, ok := httpStatusMap[errType]; ok {
		return status
	}

	// Default to 500 for unknown types
	return 500
}

// New creates a new Error with the given message and captures the stack trace
// Default error type is TypeInternal (500). Use WithType() to change it.
//
// Example:
//
//	err := lgerr.New("database connection failed").
//	    WithType(lgerr.TypeDatabase).
//	    WithContext("host", "localhost")
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
		stackTrace: pcs[:n:n], // Use slice expression to prevent unwanted mutations
		context:    make(map[string]any),
	}
}

// WithType sets the error type, which determines the default HTTP status code
func (e *Error) WithType(errType ErrorType) *Error {
	e.errorType = errType
	return e
}

// WithContext adds a key-value pair to the error's context metadata
// This metadata is included in logs and can be sent to Sentry
func (e *Error) WithContext(key string, value any) *Error {
	if e.context == nil {
		e.context = make(map[string]any)
	}
	e.context[key] = value
	return e
}

// WithHTTPStatus overrides the default HTTP status code for this specific error
func (e *Error) WithHTTPStatus(status int) *Error {
	e.httpStatus = &status
	return e
}

// Wrap wraps another error into this error
func (e *Error) Wrap(err error) *Error {
	e.wrapped = err
	return e
}

// SetHTTPStatus is a non-chainable alternative to WithHTTPStatus
func (e *Error) SetHTTPStatus(status int) {
	e.httpStatus = &status
}

// IgnoreSentry marks this error to be excluded from Sentry reporting
// Useful for expected errors or those containing sensitive information
func (e *Error) IgnoreSentry() *Error {
	e.ignoreSentry = true
	return e
}

// ShouldIgnoreSentry returns whether this error should skip Sentry reporting
func (e *Error) ShouldIgnoreSentry() bool {
	return e.ignoreSentry
}

// WithTitle sets the public-facing title (RFC 7807)
// Title is a short, human-readable summary of the problem type
func (e *Error) WithTitle(title string) *Error {
	e.title = title
	return e
}

// WithDetail sets the public-facing detail (RFC 7807)
// Detail is a human-readable explanation specific to this occurrence
func (e *Error) WithDetail(detail string) *Error {
	e.detail = detail
	return e
}


// WithValidationError adds a single validation error to the error
func (e *Error) WithValidationError(field string, message string, value ...any) *Error {
	if e.validationErrors == nil {
		e.validationErrors = make([]ValidationError, 0)
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

// Title returns the public-facing title
func (e *Error) Title() string {
	return e.title
}

// Detail returns the public-facing detail
func (e *Error) Detail() string {
	return e.detail
}


func (e *Error) ValidationErrors() []ValidationError {
	return e.validationErrors
}

// HasValidationErrors returns true if there are validation errors
func (e *Error) HasValidationErrors() bool {
	return len(e.validationErrors) > 0
}

// ToErrorResponse converts the error to an ErrorResponse (RFC 7807 format)
func (e *Error) ToErrorResponse() ErrorResponse {
	response := ErrorResponse{
		Title:  e.title,
		Detail: e.detail,
		Errors: e.validationErrors,
	}

	// Add context as meta if present
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

	var result string
	frames := runtime.CallersFrames(e.stackTrace)
	for {
		frame, more := frames.Next()
		result += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return result
}
