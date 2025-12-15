package lgfiber

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/aeternitas-infinita/logbundle-go/internal/logger"
	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"
)

// ValidationConfig holds configuration for validation middleware
type ValidationConfig struct {
	// Logger instance for validation logging
	Logger *slog.Logger
	// Validator instance (if nil, uses default validator)
	Validator *validator.Validate
	// LocalsKey is the key used to store validated data in c.Locals (default: "body", "params", etc.)
	LocalsKey string
	// Title for validation error response (default: "Validation Error")
	Title string
	// Detail for validation error response (optional)
	Detail string
}

var (
	defaultValidator     *validator.Validate
	defaultValidatorOnce sync.Once

	// Global validation middleware configs
	defaultBodyConfig    ValidationConfig
	defaultQueryConfig   ValidationConfig
	defaultParamsConfig  ValidationConfig
	defaultHeadersConfig ValidationConfig
	defaultGlobalLogger  *slog.Logger
	configMutex          sync.RWMutex

	// Pool for validation error slices to reduce allocations
	validationErrorPool = sync.Pool{
		New: func() any {
			return make([]lgerr.ValidationError, 0, 8)
		},
	}
)

// getDefaultValidator returns the default validator instance (lazy initialization)
func getDefaultValidator() *validator.Validate {
	defaultValidatorOnce.Do(func() {
		defaultValidator = validator.New()
	})
	return defaultValidator
}

// SetDefaultValidator sets a custom default validator instance
// Call this at application startup to use a custom validator with additional rules
func SetDefaultValidator(v *validator.Validate) {
	configMutex.Lock()
	defer configMutex.Unlock()
	if v != nil {
		defaultValidator = v
	}
}

// GetDefaultValidator returns the current default validator instance
func GetDefaultValidator() *validator.Validate {
	return getDefaultValidator()
}

// init initializes all default validation configs at package load time
func init() {
	defaultBodyConfig = ValidationConfig{
		LocalsKey: "body",
		Title:     "Validation Error",
		Detail:    "Please check your request body",
	}
	defaultQueryConfig = ValidationConfig{
		LocalsKey: "query",
		Title:     "Invalid Query Parameters",
		Detail:    "Please check your query parameters",
	}
	defaultParamsConfig = ValidationConfig{
		LocalsKey: "params",
		Title:     "Invalid Route Parameters",
		Detail:    "Please check your route parameters",
	}
	defaultHeadersConfig = ValidationConfig{
		LocalsKey: "headers",
		Title:     "Invalid Request Headers",
		Detail:    "Please check your request headers",
	}
}

// SetValidationLogger sets the global logger for all validation middlewares
// Call this at application startup to configure logging for validation errors
func SetValidationLogger(logger *slog.Logger) {
	configMutex.Lock()
	defaultGlobalLogger = logger
	configMutex.Unlock()
}

// GetValidationLogger returns the global validation logger
func GetValidationLogger() *slog.Logger {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return defaultGlobalLogger
}

// SetBodyValidationConfig sets the global configuration for body validation middleware
func SetBodyValidationConfig(config ValidationConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	// Keep LocalsKey and Detail as defaults if not explicitly set
	if config.Logger != nil {
		defaultBodyConfig.Logger = config.Logger
	}
	if config.Validator != nil {
		defaultBodyConfig.Validator = config.Validator
	}
	if config.Title != "" {
		defaultBodyConfig.Title = config.Title
	}
}

// GetBodyValidationConfig returns a copy of the global body validation config
func GetBodyValidationConfig() ValidationConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return defaultBodyConfig
}

// SetQueryValidationConfig sets the global configuration for query validation middleware
func SetQueryValidationConfig(config ValidationConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	// Keep LocalsKey and Detail as defaults if not explicitly set
	if config.Logger != nil {
		defaultQueryConfig.Logger = config.Logger
	}
	if config.Validator != nil {
		defaultQueryConfig.Validator = config.Validator
	}
	if config.Title != "" {
		defaultQueryConfig.Title = config.Title
	}
}

// GetQueryValidationConfig returns a copy of the global query validation config
func GetQueryValidationConfig() ValidationConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return defaultQueryConfig
}

// SetParamsValidationConfig sets the global configuration for params validation middleware
func SetParamsValidationConfig(config ValidationConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	// Keep LocalsKey and Detail as defaults if not explicitly set
	if config.Logger != nil {
		defaultParamsConfig.Logger = config.Logger
	}
	if config.Validator != nil {
		defaultParamsConfig.Validator = config.Validator
	}
	if config.Title != "" {
		defaultParamsConfig.Title = config.Title
	}
}

// GetParamsValidationConfig returns a copy of the global params validation config
func GetParamsValidationConfig() ValidationConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return defaultParamsConfig
}

// SetHeadersValidationConfig sets the global configuration for headers validation middleware
func SetHeadersValidationConfig(config ValidationConfig) {
	configMutex.Lock()
	defer configMutex.Unlock()
	// Keep LocalsKey and Detail as defaults if not explicitly set
	if config.Logger != nil {
		defaultHeadersConfig.Logger = config.Logger
	}
	if config.Validator != nil {
		defaultHeadersConfig.Validator = config.Validator
	}
	if config.Title != "" {
		defaultHeadersConfig.Title = config.Title
	}
}

// GetHeadersValidationConfig returns a copy of the global headers validation config
func GetHeadersValidationConfig() ValidationConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return defaultHeadersConfig
}

// ResetValidationConfigs resets all validation configs to their defaults
func ResetValidationConfigs() {
	configMutex.Lock()
	defer configMutex.Unlock()
	defaultGlobalLogger = nil
	defaultValidator = nil

	// Re-initialize to defaults
	defaultBodyConfig = ValidationConfig{
		LocalsKey: "body",
		Title:     "Validation Error",
		Detail:    "Please check your request body",
	}
	defaultQueryConfig = ValidationConfig{
		LocalsKey: "query",
		Title:     "Invalid Query Parameters",
		Detail:    "Please check your query parameters",
	}
	defaultParamsConfig = ValidationConfig{
		LocalsKey: "params",
		Title:     "Invalid Route Parameters",
		Detail:    "Please check your route parameters",
	}
	defaultHeadersConfig = ValidationConfig{
		LocalsKey: "headers",
		Title:     "Invalid Request Headers",
		Detail:    "Please check your request headers",
	}
}

// parseValidationErrors converts validator.ValidationErrors to lgerr.ValidationError slice
func parseValidationErrors(err error, dto interface{}) []lgerr.ValidationError {
	if validatorErrs, ok := err.(validator.ValidationErrors); ok {
		// Get slice from pool and reset it
		validationErrors := validationErrorPool.Get().([]lgerr.ValidationError)
		validationErrors = validationErrors[:0]

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

		// Make a copy to return (pool will be reused)
		result := make([]lgerr.ValidationError, len(validationErrors))
		copy(result, validationErrors)

		// Return slice to pool
		validationErrorPool.Put(validationErrors)

		return result
	}

	return nil
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
