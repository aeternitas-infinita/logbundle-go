package lgerr

import "fmt"

func NotFound(resource string, id any) *Error {
	return New(fmt.Sprintf("%s not found", resource)).
		WithType(TypeNotFound).
		WithContext("resource", resource).
		WithContext("resource_id", id)
}

func Validation(field string, reason string) *Error {
	return New(fmt.Sprintf("validation failed: %s", reason)).
		WithType(TypeValidation).
		WithContext("field", field).
		WithContext("reason", reason)
}

func Database(message string) *Error {
	return New(message).WithType(TypeDatabase)
}

func Internal(message string) *Error {
	return New(message).WithType(TypeInternal)
}

func Forbidden(resource string, id any, reason string) *Error {
	return New(fmt.Sprintf("access forbidden: %s", reason)).
		WithType(TypeForbidden).
		WithContext("resource", resource).
		WithContext("resource_id", id).
		WithContext("reason", reason)
}

func Unauthorized(reason string) *Error {
	return New(fmt.Sprintf("unauthorized: %s", reason)).
		WithType(TypeUnauth).
		WithContext("reason", reason)
}

func BadInput(message string) *Error {
	return New(message).WithType(TypeBadInput)
}

func Conflict(resource, field string, value any, reason string) *Error {
	return New(fmt.Sprintf("%s conflict: %s", resource, reason)).
		WithType(TypeConflict).
		WithContext("resource", resource).
		WithContext("field", field).
		WithContext("value", value).
		WithContext("reason", reason)
}

func External(service, message string) *Error {
	return New(fmt.Sprintf("external service error: %s - %s", service, message)).
		WithType(TypeExternal).
		WithContext("service", service)
}

func Timeout(operation, duration string) *Error {
	return New(fmt.Sprintf("timeout: %s exceeded %s", operation, duration)).
		WithType(TypeTimeout).
		WithContext("operation", operation).
		WithContext("duration", duration)
}

func Busy(message string) *Error {
	return New(message).WithType(TypeBusy)
}
