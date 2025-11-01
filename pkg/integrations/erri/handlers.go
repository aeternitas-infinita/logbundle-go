package erri

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/pkg/core"
	"github.com/aeternitas-infinita/logbundle-go/pkg/handler"
)

// Handle processes an error and returns appropriate HTTP status and response
// It logs internal errors and database errors, and formats user-facing responses
func Handle(ctx context.Context, err error, c *fiber.Ctx) (int, *HttpResponse) {
	var internalErr *Erri
	if errors.As(err, &internalErr) {
		statusCode := internalErr.HTTPStatusCode()

		// Log severe errors (5xx and database errors)
		if statusCode == http.StatusInternalServerError ||
			internalErr.Type == ErriStruct.DATABASE {
			requestInfo := extractRequestInfo(c)

			handler.Log.ErrorContext(
				ctx,
				"Handled internal error",
				core.ErrAttr(internalErr),
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

		// Return structured response if property and message are set
		if internalErr.Property == "" || internalErr.Message == "" {
			return statusCode, &HttpResponse{
				Message: "Oops, something went wrong",
			}
		}
		return statusCode, &HttpResponse{
			AnswerInfo: []AnswerInfoType{{Property: internalErr.Property, Message: internalErr.Message}},
		}
	}

	// Handle non-Erri errors
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

// LogErri logs an Erri error with full context using a custom logger
func LogErri(ctx context.Context, internalErr *Erri, logger *slog.Logger, c *fiber.Ctx) {
	var requestInfo requestInfo
	if c != nil {
		requestInfo = extractRequestInfo(c)
	}

	logger.ErrorContext(
		ctx,
		"Logged internal error",
		core.ErrAttr(internalErr),
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

// extractRequestInfo extracts HTTP request information from Fiber context
func extractRequestInfo(c *fiber.Ctx) requestInfo {
	var params map[string]any
	if paramsValue := c.Locals("params"); paramsValue != nil {
		params = map[string]any{
			"params": paramsValue,
		}
	}

	queryParams := make(map[string]any)
	for key, value := range c.Context().QueryArgs().All() {
		queryParams[string(key)] = string(value)
	}

	return requestInfo{
		URL:         c.OriginalURL(),
		Method:      c.Method(),
		Params:      params,
		QueryParams: queryParams,
		Route:       c.Route().Path,
	}
}
