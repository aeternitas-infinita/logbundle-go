package erri

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
	"github.com/aeternitas-infinita/logbundle-go/pkg/handler"
)

type ErriType string

type requestInfo struct {
	URL         string         `json:"url"`
	Method      string         `json:"method"`
	Params      map[string]any `json:"params,omitempty"`
	QueryParams map[string]any `json:"query_params,omitempty"`
	Route       string         `json:"route"`
}

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

type Erri struct {
	Type        ErriType
	Property    string
	Value       any
	Message     string
	Details     string
	File        string
	SystemError error
}

func (e *Erri) Error() string {
	return fmt.Sprintf("handled internal error. Details: '%s', file: '%s', type: '%s' system error: '%v'", e.Details, e.File, e.Type, e.SystemError)
}

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

func New() *ErriBuilder {
	return &ErriBuilder{
		err: &Erri{
			File: core.GetLinePositionStringWithSkip(2),
		},
	}
}

func Handle(ctx context.Context, err error, c *fiber.Ctx) (int, *HttpResponse) {
	var internalErr *Erri
	if errors.As(err, &internalErr) {
		statusCode := internalErr.HTTPStatusCode()

		if statusCode == http.StatusInternalServerError ||
			internalErr.Type == ErriStruct.DATABASE {
			requestInfo := extractRequestInfo(c)

			handler.Log.ErrorContext(
				ctx,
				"Handled internal error",
				slog.String("details", internalErr.Details),
				slog.String("file", internalErr.File),
				slog.String("message", internalErr.Message),
				slog.Any("value", internalErr.Value),
				slog.String("property", internalErr.Property),
				slog.String("type", string(internalErr.Type)),
				slog.Any("system_error", internalErr.SystemError),
				slog.String("request_url", requestInfo.URL),
				slog.String("request_method", requestInfo.Method),
				slog.String("request_route", requestInfo.Route),
				slog.Any("request_params", requestInfo.Params),
				slog.Any("request_query_params", requestInfo.QueryParams),
			)
		}

		if internalErr.Property == "" || internalErr.Message == "" {
			return statusCode, &HttpResponse{
				Message: "Oops, something went wrong",
			}
		}
		return statusCode, &HttpResponse{
			AnswerInfo: []AnswerInfoType{{Property: internalErr.Property, Message: internalErr.Message}},
		}
	}

	if c != nil {
		requestInfo := extractRequestInfo(c)
		handler.Log.ErrorContext(ctx, "handled error",
			core.ErrAttr(err),
			slog.String("request_url", requestInfo.URL),
			slog.String("request_method", requestInfo.Method),
			slog.String("request_route", requestInfo.Route),
			slog.Any("request_params", requestInfo.Params),
			slog.Any("request_query_params", requestInfo.QueryParams),
		)
	} else {
		handler.Log.ErrorContext(ctx, "handled error", core.ErrAttr(err))
	}

	return http.StatusInternalServerError, nil
}

type AnswerInfoType struct {
	Property string `json:"property,omitempty"`
	CodeType int    `json:"code_type,omitempty"`
	Message  string `json:"message,omitempty"`
}

type HttpResponse struct {
	Data       any              `json:"data,omitempty"`
	AnswerCode int              `json:"answer_code,omitempty"`
	AnswerInfo []AnswerInfoType `json:"answer_info,omitempty"`
	Message    string           `json:"message,omitempty"`
}

func (mr *HttpResponse) Error() string {
	return fmt.Sprintf("Message: %s", mr.Message)
}

func LogErri(ctx context.Context, internalErr *Erri, logger *slog.Logger, c *fiber.Ctx) {
	var requestInfo requestInfo
	if c != nil {
		requestInfo = extractRequestInfo(c)
	}

	logger.ErrorContext(
		ctx,
		"Logged internal error",
		slog.String("details", internalErr.Details),
		slog.String("file", internalErr.File),
		slog.String("message", internalErr.Message),
		slog.Any("value", internalErr.Value),
		slog.String("property", internalErr.Property),
		slog.String("type", string(internalErr.Type)),
		slog.Any("system_error", internalErr.SystemError),
		slog.String("request_url", requestInfo.URL),
		slog.String("request_method", requestInfo.Method),
		slog.String("request_route", requestInfo.Route),
		slog.Any("request_params", requestInfo.Params),
		slog.Any("request_query_params", requestInfo.QueryParams),
	)
}
