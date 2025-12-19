package lgfiber

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/aeternitas-infinita/logbundle-go/internal/logger"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"
	"github.com/gofiber/fiber/v2"
)

// genericValidationMiddleware creates a validation middleware for any parser function
func genericValidationMiddleware[T any](
	parserFunc func(*fiber.Ctx, *T) error,
	config ValidationConfig,
) fiber.Handler {
	// Set defaults
	if config.Validator == nil {
		config.Validator = getDefaultValidator()
	}
	if config.Title == "" {
		config.Title = "Validation Error"
	}

	return func(c *fiber.Ctx) error {
		var dto T

		// Parse the request
		if err := parserFunc(c, &dto); err != nil {
			if config.Logger != nil {
				logger.LogWithSourceCtx(c.UserContext(), config.Logger, slog.LevelWarn, "Failed to parse request",
					"error", err.Error(),
					"parser", config.LocalsKey,
				)
			}

			return c.Status(http.StatusBadRequest).JSON(lgerr.ErrorResponse{
				Title:  "Invalid Request Format",
				Detail: "Failed to parse request: " + err.Error(),
			})
		}

		// Validate the parsed data
		if err := config.Validator.Struct(dto); err != nil {
			validationErrors := parseValidationErrors(err, dto)

			if len(validationErrors) > 0 {
				if config.Logger != nil {
					logger.LogWithSourceCtx(c.UserContext(), config.Logger, slog.LevelDebug, "Validation failed",
						"errors_count", len(validationErrors),
						"parser", config.LocalsKey,
					)
				}

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
// Uses the global body validation config set via SetBodyValidationConfig()
//
// Usage:
//
//	type CreateUserRequest struct {
//	    Email string `json:"email" validate:"required,email"`
//	    Name  string `json:"name" validate:"required,min=2,max=100"`
//	}
//
//	// At startup: configure globally
//	lgfiber.SetValidationLogger(appLogger)
//	lgfiber.SetBodyValidationConfig(lgfiber.ValidationConfig{
//	    Title: "Invalid User Request",
//	})
//
//	// In routes: use global config
//	app.Post("/users", lgfiber.BodyValidationMiddleware[CreateUserRequest](), handler)
//
//	func handler(c *fiber.Ctx) error {
//	    body := c.Locals("body").(CreateUserRequest)
//	    // Use validated body...
//	}
func BodyValidationMiddleware[T any]() fiber.Handler {
	// Capture global config once at middleware creation (not per-request)
	configMutex.RLock()
	config := defaultBodyConfig
	if defaultGlobalLogger != nil && config.Logger == nil {
		config.Logger = defaultGlobalLogger
	}
	configMutex.RUnlock()

	return genericValidationMiddleware(
		func(ctx *fiber.Ctx, dto *T) error { return ctx.BodyParser(dto) },
		config,
	)
}

// QueryValidationMiddleware creates a middleware that validates query parameters
// Uses the global query validation config set via SetQueryValidationConfig()
//
// Usage:
//
//	type SearchQuery struct {
//	    Query string `json:"query" validate:"required,min=3"`
//	    Limit int    `json:"limit" validate:"min=1,max=100"`
//	}
//
//	// At startup: configure globally
//	lgfiber.SetValidationLogger(appLogger)
//	lgfiber.SetQueryValidationConfig(lgfiber.ValidationConfig{
//	    Title: "Invalid search query",
//	})
//
//	// In routes: use global config
//	app.Get("/search", lgfiber.QueryValidationMiddleware[SearchQuery](), handler)
//
//	func handler(c *fiber.Ctx) error {
//	    query := c.Locals("query").(SearchQuery)
//	    // Use validated query...
//	}
func QueryValidationMiddleware[T any]() fiber.Handler {
	// Capture global config once at middleware creation (not per-request)
	configMutex.RLock()
	config := defaultQueryConfig
	if defaultGlobalLogger != nil && config.Logger == nil {
		config.Logger = defaultGlobalLogger
	}
	configMutex.RUnlock()

	return genericValidationMiddleware(
		func(ctx *fiber.Ctx, dto *T) error { return ctx.QueryParser(dto) },
		config,
	)
}

// ParamsValidationMiddleware creates a middleware that validates route parameters
// Uses the global params validation config set via SetParamsValidationConfig()
//
// Usage:
//
//	type UserParams struct {
//	    ID string `params:"id" validate:"required,uuid"`
//	}
//
//	// At startup: configure globally
//	lgfiber.SetValidationLogger(appLogger)
//	lgfiber.SetParamsValidationConfig(lgfiber.ValidationConfig{
//	    Title: "Invalid user ID",
//	})
//
//	// In routes: use global config
//	app.Get("/users/:id", lgfiber.ParamsValidationMiddleware[UserParams](), handler)
//
//	func handler(c *fiber.Ctx) error {
//	    params := c.Locals("params").(UserParams)
//	    // Use validated params...
//	}
func ParamsValidationMiddleware[T any]() fiber.Handler {
	// Capture global config once at middleware creation (not per-request)
	configMutex.RLock()
	config := defaultParamsConfig
	if defaultGlobalLogger != nil && config.Logger == nil {
		config.Logger = defaultGlobalLogger
	}
	configMutex.RUnlock()

	return genericValidationMiddleware(
		func(ctx *fiber.Ctx, dto *T) error { return ctx.ParamsParser(dto) },
		config,
	)
}

// HeadersValidationMiddleware creates a middleware that validates request headers
// Uses the global headers validation config set via SetHeadersValidationConfig()
//
// Usage:
//
//	type RequiredHeaders struct {
//	    Authorization string `reqheader:"Authorization" validate:"required"`
//	    ContentType   string `reqheader:"Content-Type" validate:"required"`
//	}
//
//	// At startup: configure globally
//	lgfiber.SetValidationLogger(appLogger)
//	lgfiber.SetHeadersValidationConfig(lgfiber.ValidationConfig{
//	    Title: "Missing required headers",
//	})
//
//	// In routes: use global config
//	app.Post("/api", lgfiber.HeadersValidationMiddleware[RequiredHeaders](), handler)
//
//	func handler(c *fiber.Ctx) error {
//	    headers := c.Locals("headers").(RequiredHeaders)
//	    // Use validated headers...
//	}
func HeadersValidationMiddleware[T any]() fiber.Handler {
	// Capture global config once at middleware creation (not per-request)
	configMutex.RLock()
	config := defaultHeadersConfig
	if defaultGlobalLogger != nil && config.Logger == nil {
		config.Logger = defaultGlobalLogger
	}
	configMutex.RUnlock()

	return genericValidationMiddleware(
		func(ctx *fiber.Ctx, dto *T) error { return ctx.ReqHeaderParser(dto) },
		config,
	)
}

// FormDataValidationMiddleware creates a middleware that validates form data with JSON in a specific field
// Expects form data with a field containing JSON that will be validated
// Uses the global body validation config set via SetBodyValidationConfig()
//
// Usage:
//
//	type CreateUserRequest struct {
//	    Email string `json:"email" validate:"required,email"`
//	    Name  string `json:"name" validate:"required,min=2,max=100"`
//	}
//
//	// At startup: configure globally
//	lgfiber.SetValidationLogger(appLogger)
//	lgfiber.SetBodyValidationConfig(lgfiber.ValidationConfig{
//	    Title: "Invalid Form Data",
//	})
//
//	// In routes: use global config (defaults to "json_data" field)
//	app.Post("/users", lgfiber.FormDataValidationMiddleware[CreateUserRequest](""), handler)
//
//	// Or with custom field name
//	app.Post("/users", lgfiber.FormDataValidationMiddleware[CreateUserRequest]("data"), handler)
//
//	func handler(c *fiber.Ctx) error {
//	    body := c.Locals("form_data").(CreateUserRequest)
//	    // Use validated body...
//	}
func FormDataValidationMiddleware[T any](formFieldName string) fiber.Handler {
	fieldName := "json_data"
	if formFieldName != "" {
		fieldName = formFieldName
	}

	// Capture global config once at middleware creation (not per-request)
	configMutex.RLock()
	config := ValidationConfig{
		Logger:    defaultBodyConfig.Logger,
		Validator: defaultBodyConfig.Validator,
		Title:     defaultBodyConfig.Title,
		Detail:    defaultBodyConfig.Detail,
		LocalsKey: "form_data",
	}
	if defaultGlobalLogger != nil && config.Logger == nil {
		config.Logger = defaultGlobalLogger
	}
	configMutex.RUnlock()

	return genericValidationMiddleware(
		func(ctx *fiber.Ctx, dto *T) error {
			// Get form value
			bodyStr := ctx.FormValue(fieldName)
			if bodyStr == "" {
				return fiber.NewError(fiber.StatusBadRequest, "missing form field: "+fieldName)
			}

			// Unmarshal JSON from form field
			if err := json.Unmarshal([]byte(bodyStr), dto); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "invalid JSON in form field: "+err.Error())
			}

			return nil
		},
		config,
	)
}
