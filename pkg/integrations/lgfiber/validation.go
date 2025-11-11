package lgfiber

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"
)

// ValidationConfig holds configuration for validation middleware
type ValidationConfig struct {
	// Validator instance (if nil, uses default validator)
	Validator *validator.Validate
	// LocalsKey is the key used to store validated data in c.Locals (default: "body", "params", etc.)
	LocalsKey string
	// Title for validation error response (default: "Validation Error")
	Title string
	// Detail for validation error response (optional)
	Detail string
}

var defaultValidator = validator.New()

// parseValidationErrors converts validator.ValidationErrors to lgerr.ValidationError slice
func parseValidationErrors(err error, dto interface{}) []lgerr.ValidationError {
	var validationErrors []lgerr.ValidationError

	if validatorErrs, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validatorErrs {
			// Get the JSON field name from struct tag
			fieldName := getJSONFieldName(dto, fieldErr.Field())
			if fieldName == "" {
				fieldName = strings.ToLower(fieldErr.Field())
			}

			validationErr := lgerr.ValidationError{
				Field:   fieldName,
				Message: getValidationMessage(fieldErr),
				Value:   fieldErr.Value(),
			}
			validationErrors = append(validationErrors, validationErr)
		}
	}

	return validationErrors
}

// getJSONFieldName extracts the JSON field name from struct tag
func getJSONFieldName(dto interface{}, fieldName string) string {
	t := reflect.TypeOf(dto)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return ""
	}

	field, found := t.FieldByName(fieldName)
	if !found {
		return ""
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return ""
	}

	// Extract field name from tag (before comma)
	parts := strings.Split(jsonTag, ",")
	if parts[0] == "-" {
		return ""
	}

	return parts[0]
}

// getValidationMessage returns a human-readable error message for the validation tag
func getValidationMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short or small (min: " + fieldErr.Param() + ")"
	case "max":
		return "Value is too long or large (max: " + fieldErr.Param() + ")"
	case "len":
		return "Value must have length of " + fieldErr.Param()
	case "gt":
		return "Value must be greater than " + fieldErr.Param()
	case "gte":
		return "Value must be greater than or equal to " + fieldErr.Param()
	case "lt":
		return "Value must be less than " + fieldErr.Param()
	case "lte":
		return "Value must be less than or equal to " + fieldErr.Param()
	case "url":
		return "Invalid URL format"
	case "uuid":
		return "Invalid UUID format"
	case "alpha":
		return "Only alphabetic characters allowed"
	case "alphanum":
		return "Only alphanumeric characters allowed"
	case "numeric":
		return "Only numeric characters allowed"
	case "oneof":
		return "Value must be one of: " + fieldErr.Param()
	default:
		return "Validation failed: " + fieldErr.Tag()
	}
}

// genericValidationMiddleware creates a validation middleware for any parser function
func genericValidationMiddleware[T any](
	parserFunc func(*fiber.Ctx, *T) error,
	config ValidationConfig,
) fiber.Handler {
	// Set defaults
	if config.Validator == nil {
		config.Validator = defaultValidator
	}
	if config.Title == "" {
		config.Title = "Validation Error"
	}

	return func(c *fiber.Ctx) error {
		var dto T

		// Parse the request
		if err := parserFunc(c, &dto); err != nil {
			logbundle.WarnCtx(c.UserContext(), "Failed to parse request",
				"error", err.Error(),
				"parser", config.LocalsKey,
			)

			return c.Status(http.StatusBadRequest).JSON(lgerr.ErrorResponse{
				Title:  "Invalid Request Format",
				Detail: "Failed to parse request: " + err.Error(),
			})
		}

		// Validate the parsed data
		if err := config.Validator.Struct(dto); err != nil {
			validationErrors := parseValidationErrors(err, dto)

			if len(validationErrors) > 0 {
				logbundle.DebugCtx(c.UserContext(), "Validation failed",
					"errors_count", len(validationErrors),
					"parser", config.LocalsKey,
				)

				response := lgerr.ErrorResponse{
					Title:  config.Title,
					Errors: validationErrors,
				}

				if config.Detail != "" {
					response.Detail = config.Detail
				}

				return c.Status(http.StatusUnprocessableEntity).JSON(response)
			}
		}

		// Store validated data in locals
		c.Locals(config.LocalsKey, dto)
		return c.Next()
	}
}

// BodyValidationMiddleware creates a middleware that validates request body
// Usage:
//
//	type CreateUserRequest struct {
//	    Email string `json:"email" validate:"required,email"`
//	    Name  string `json:"name" validate:"required,min=2,max=100"`
//	}
//
//	app.Post("/users", lgfiber.BodyValidationMiddleware[CreateUserRequest](), handler)
//
//	func handler(c *fiber.Ctx) error {
//	    body := c.Locals("body").(CreateUserRequest)
//	    // Use validated body...
//	}
func BodyValidationMiddleware[T any](customConfig ...ValidationConfig) fiber.Handler {
	config := ValidationConfig{
		LocalsKey: "body",
		Title:     "Validation Error",
		Detail:    "Please check your request body",
	}

	if len(customConfig) > 0 {
		if customConfig[0].Validator != nil {
			config.Validator = customConfig[0].Validator
		}
		if customConfig[0].LocalsKey != "" {
			config.LocalsKey = customConfig[0].LocalsKey
		}
		if customConfig[0].Title != "" {
			config.Title = customConfig[0].Title
		}
		if customConfig[0].Detail != "" {
			config.Detail = customConfig[0].Detail
		}
	}

	return genericValidationMiddleware[T](
		func(ctx *fiber.Ctx, dto *T) error { return ctx.BodyParser(dto) },
		config,
	)
}

// QueryValidationMiddleware creates a middleware that validates query parameters
// Usage:
//
//	type SearchQuery struct {
//	    Query string `json:"query" validate:"required,min=3"`
//	    Limit int    `json:"limit" validate:"min=1,max=100"`
//	}
//
//	app.Get("/search", lgfiber.QueryValidationMiddleware[SearchQuery](), handler)
//
//	func handler(c *fiber.Ctx) error {
//	    query := c.Locals("query").(SearchQuery)
//	    // Use validated query...
//	}
func QueryValidationMiddleware[T any](customConfig ...ValidationConfig) fiber.Handler {
	config := ValidationConfig{
		LocalsKey: "query",
		Title:     "Invalid Query Parameters",
		Detail:    "Please check your query parameters",
	}

	if len(customConfig) > 0 {
		if customConfig[0].Validator != nil {
			config.Validator = customConfig[0].Validator
		}
		if customConfig[0].LocalsKey != "" {
			config.LocalsKey = customConfig[0].LocalsKey
		}
		if customConfig[0].Title != "" {
			config.Title = customConfig[0].Title
		}
		if customConfig[0].Detail != "" {
			config.Detail = customConfig[0].Detail
		}
	}

	return genericValidationMiddleware[T](
		func(ctx *fiber.Ctx, dto *T) error { return ctx.QueryParser(dto) },
		config,
	)
}

// ParamsValidationMiddleware creates a middleware that validates route parameters
// Usage:
//
//	type UserParams struct {
//	    ID string `params:"id" validate:"required,uuid"`
//	}
//
//	app.Get("/users/:id", lgfiber.ParamsValidationMiddleware[UserParams](), handler)
//
//	func handler(c *fiber.Ctx) error {
//	    params := c.Locals("params").(UserParams)
//	    // Use validated params...
//	}
func ParamsValidationMiddleware[T any](customConfig ...ValidationConfig) fiber.Handler {
	config := ValidationConfig{
		LocalsKey: "params",
		Title:     "Invalid Route Parameters",
		Detail:    "Please check your route parameters",
	}

	if len(customConfig) > 0 {
		if customConfig[0].Validator != nil {
			config.Validator = customConfig[0].Validator
		}
		if customConfig[0].LocalsKey != "" {
			config.LocalsKey = customConfig[0].LocalsKey
		}
		if customConfig[0].Title != "" {
			config.Title = customConfig[0].Title
		}
		if customConfig[0].Detail != "" {
			config.Detail = customConfig[0].Detail
		}
	}

	return genericValidationMiddleware[T](
		func(ctx *fiber.Ctx, dto *T) error { return ctx.ParamsParser(dto) },
		config,
	)
}

// HeadersValidationMiddleware creates a middleware that validates request headers
// Usage:
//
//	type RequiredHeaders struct {
//	    Authorization string `reqheader:"Authorization" validate:"required"`
//	    ContentType   string `reqheader:"Content-Type" validate:"required"`
//	}
//
//	app.Post("/api", lgfiber.HeadersValidationMiddleware[RequiredHeaders](), handler)
//
//	func handler(c *fiber.Ctx) error {
//	    headers := c.Locals("headers").(RequiredHeaders)
//	    // Use validated headers...
//	}
func HeadersValidationMiddleware[T any](customConfig ...ValidationConfig) fiber.Handler {
	config := ValidationConfig{
		LocalsKey: "headers",
		Title:     "Invalid Request Headers",
		Detail:    "Please check your request headers",
	}

	if len(customConfig) > 0 {
		if customConfig[0].Validator != nil {
			config.Validator = customConfig[0].Validator
		}
		if customConfig[0].LocalsKey != "" {
			config.LocalsKey = customConfig[0].LocalsKey
		}
		if customConfig[0].Title != "" {
			config.Title = customConfig[0].Title
		}
		if customConfig[0].Detail != "" {
			config.Detail = customConfig[0].Detail
		}
	}

	return genericValidationMiddleware[T](
		func(ctx *fiber.Ctx, dto *T) error { return ctx.ReqHeaderParser(dto) },
		config,
	)
}
