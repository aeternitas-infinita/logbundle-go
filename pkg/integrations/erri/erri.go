// Package erri provides structured error handling for Go applications with HTTP status code mapping.
// It includes a builder pattern for creating detailed errors and integration with Fiber for HTTP responses.
package erri

import (
	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
)

// New creates a new error builder with automatic file/line tracking
func New() *ErriBuilder {
	return &ErriBuilder{
		err: &Erri{
			File: core.GetLinePositionStringWithSkip(2),
		},
	}
}
