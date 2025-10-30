package erri

// ErriBuilder provides a fluent interface for building Erri errors
type ErriBuilder struct {
	err *Erri
}

// Type sets the error type
func (b *ErriBuilder) Type(errorType ErriType) *ErriBuilder {
	b.err.Type = errorType
	return b
}

// Message sets the user-facing error message
func (b *ErriBuilder) Message(message string) *ErriBuilder {
	b.err.Message = message
	return b
}

// Details sets additional error details for logging
func (b *ErriBuilder) Details(details string) *ErriBuilder {
	b.err.Details = details
	return b
}

// Property sets the property/field that caused the error
func (b *ErriBuilder) Property(property string) *ErriBuilder {
	b.err.Property = property
	return b
}

// Value sets the value that caused the error
func (b *ErriBuilder) Value(value any) *ErriBuilder {
	b.err.Value = value
	return b
}

// SystemError sets the underlying system error
func (b *ErriBuilder) SystemError(systemError error) *ErriBuilder {
	b.err.SystemError = systemError
	return b
}

// Build returns the constructed Erri error
func (b *ErriBuilder) Build() *Erri {
	return b.err
}
