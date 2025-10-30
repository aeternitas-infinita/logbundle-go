package erri

import (
	"fmt"
	"net/http"
)

// ErriType represents the category of error
type ErriType string

// ErriStruct contains all predefined error types
var ErriStruct = struct {
	NOT_FOUND   ErriType
	VALIDATION  ErriType
	DATABASE    ErriType
	INTERNAL    ErriType
	BUSY        ErriType
	FORBIDDEN   ErriType
	WRONG_INPUT ErriType
}{
	NOT_FOUND:   "NOT_FOUND",
	VALIDATION:  "VALIDATION",
	DATABASE:    "DATABASE",
	INTERNAL:    "INTERNAL",
	BUSY:        "BUSY",
	FORBIDDEN:   "FORBIDDEN",
	WRONG_INPUT: "WRONG_INPUT",
}

// Erri represents a structured internal error with rich context
type Erri struct {
	Type        ErriType
	Property    string
	Value       any
	Message     string
	Details     string
	File        string
	SystemError error
}

// Error implements the error interface
func (e *Erri) Error() string {
	return fmt.Sprintf("handled internal error. Details: '%s', file: '%s', type: '%s' system error: '%v'",
		e.Details, e.File, e.Type, e.SystemError)
}

// HTTPStatusCode maps error type to HTTP status code
func (e *Erri) HTTPStatusCode() int {
	switch e.Type {
	case ErriStruct.NOT_FOUND:
		return http.StatusNotFound
	case ErriStruct.VALIDATION:
		return http.StatusBadRequest
	case ErriStruct.DATABASE:
		return http.StatusInternalServerError
	case ErriStruct.INTERNAL:
		return http.StatusInternalServerError
	case ErriStruct.FORBIDDEN:
		return http.StatusForbidden
	case ErriStruct.BUSY:
		return http.StatusConflict
	case ErriStruct.WRONG_INPUT:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// AnswerInfoType represents structured error information for API responses
type AnswerInfoType struct {
	Property string `json:"property,omitempty"`
	CodeType int    `json:"code_type,omitempty"`
	Message  string `json:"message,omitempty"`
}

// HttpResponse represents a standardized HTTP error response
type HttpResponse struct {
	Data       any              `json:"data,omitempty"`
	AnswerCode int              `json:"answer_code,omitempty"`
	AnswerInfo []AnswerInfoType `json:"answer_info,omitempty"`
	Message    string           `json:"message,omitempty"`
}

// Error implements the error interface for HttpResponse
func (mr *HttpResponse) Error() string {
	return fmt.Sprintf("Message: %s", mr.Message)
}

// requestInfo contains HTTP request information for logging
type requestInfo struct {
	URL         string         `json:"url"`
	Method      string         `json:"method"`
	Params      map[string]any `json:"params,omitempty"`
	QueryParams map[string]any `json:"query_params,omitempty"`
	Route       string         `json:"route"`
}
