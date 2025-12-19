package lgfiber

import (
	"log/slog"
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
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

	defaultBodyConfig    ValidationConfig
	defaultQueryConfig   ValidationConfig
	defaultParamsConfig  ValidationConfig
	defaultHeadersConfig ValidationConfig
	defaultGlobalLogger  *slog.Logger
	configMutex          sync.RWMutex

	fieldNameCache      = make(map[reflect.Type]map[string]string)
	fieldNameCacheMutex sync.RWMutex
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
