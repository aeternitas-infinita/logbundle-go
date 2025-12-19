package lgfiber

import (
	"reflect"
	"strings"

	"github.com/aeternitas-infinita/logbundle-go/pkg/integrations/lgerr"
	"github.com/go-playground/validator/v10"
)

// parseValidationErrors converts validator.ValidationErrors to lgerr.ValidationError slice
func parseValidationErrors(err error, dto any) []lgerr.ValidationError {
	if validatorErrs, ok := err.(validator.ValidationErrors); ok {
		validationErrors := make([]lgerr.ValidationError, 0, len(validatorErrs))

		for _, fieldErr := range validatorErrs {
			fieldName := getJSONFieldName(dto, fieldErr.Field())
			if fieldName == "" {
				fieldName = strings.ToLower(fieldErr.Field())
			}

			validationErrors = append(validationErrors, lgerr.ValidationError{
				Field:   fieldName,
				Message: getValidationMessage(fieldErr),
				Value:   fieldErr.Value(),
			})
		}

		return validationErrors
	}

	return nil
}

// getJSONFieldName extracts JSON field name from struct field with reflection caching
func getJSONFieldName(dto any, fieldName string) string {
	t := reflect.TypeOf(dto)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return ""
	}

	fieldNameCacheMutex.RLock()
	if typeCache, exists := fieldNameCache[t]; exists {
		if jsonName, found := typeCache[fieldName]; found {
			fieldNameCacheMutex.RUnlock()
			return jsonName
		}
	}
	fieldNameCacheMutex.RUnlock()

	field, found := t.FieldByName(fieldName)
	if !found {
		return ""
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return ""
	}

	parts := strings.Split(jsonTag, ",")
	if parts[0] == "-" {
		return ""
	}

	jsonName := parts[0]

	fieldNameCacheMutex.Lock()
	// Prevent unbounded cache growth - only cache if under limit
	if len(fieldNameCache) < cacheMaxSize {
		if fieldNameCache[t] == nil {
			fieldNameCache[t] = make(map[string]string)
		}
		fieldNameCache[t][fieldName] = jsonName
	}
	fieldNameCacheMutex.Unlock()

	return jsonName
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
