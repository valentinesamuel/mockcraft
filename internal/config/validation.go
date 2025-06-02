package config

import (
	"fmt"
	"reflect"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validate validates a configuration value
func Validate(value interface{}, rules map[string]interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Check required fields
	if required, ok := rules["required"].(bool); ok && required {
		if val.IsZero() {
			return &ValidationError{
				Field:   "value",
				Message: "field is required",
			}
		}
	}

	// Check type-specific validations
	switch val.Kind() {
	case reflect.String:
		return validateString(val.String(), rules)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return validateInt(val.Int(), rules)
	case reflect.Float32, reflect.Float64:
		return validateFloat(val.Float(), rules)
	case reflect.Slice:
		return validateSlice(val, rules)
	case reflect.Map:
		return validateMap(val, rules)
	}

	return nil
}

// validateString validates a string value
func validateString(value string, rules map[string]interface{}) error {
	// Check min length
	if min, ok := rules["min_length"].(int); ok {
		if len(value) < min {
			return &ValidationError{
				Field:   "value",
				Message: fmt.Sprintf("length must be at least %d", min),
			}
		}
	}

	// Check max length
	if max, ok := rules["max_length"].(int); ok {
		if len(value) > max {
			return &ValidationError{
				Field:   "value",
				Message: fmt.Sprintf("length must be at most %d", max),
			}
		}
	}

	// Check pattern
	if _, ok := rules["pattern"].(string); ok {
		// TODO: Implement regex pattern matching
	}

	return nil
}

// validateInt validates an integer value
func validateInt(value int64, rules map[string]interface{}) error {
	// Check min value
	if min, ok := rules["min"].(int64); ok {
		if value < min {
			return &ValidationError{
				Field:   "value",
				Message: fmt.Sprintf("must be at least %d", min),
			}
		}
	}

	// Check max value
	if max, ok := rules["max"].(int64); ok {
		if value > max {
			return &ValidationError{
				Field:   "value",
				Message: fmt.Sprintf("must be at most %d", max),
			}
		}
	}

	return nil
}

// validateFloat validates a float value
func validateFloat(value float64, rules map[string]interface{}) error {
	// Check min value
	if min, ok := rules["min"].(float64); ok {
		if value < min {
			return &ValidationError{
				Field:   "value",
				Message: fmt.Sprintf("must be at least %f", min),
			}
		}
	}

	// Check max value
	if max, ok := rules["max"].(float64); ok {
		if value > max {
			return &ValidationError{
				Field:   "value",
				Message: fmt.Sprintf("must be at most %f", max),
			}
		}
	}

	return nil
}

// validateSlice validates a slice value
func validateSlice(value reflect.Value, rules map[string]interface{}) error {
	// Check min length
	if min, ok := rules["min_items"].(int); ok {
		if value.Len() < min {
			return &ValidationError{
				Field:   "value",
				Message: fmt.Sprintf("must have at least %d items", min),
			}
		}
	}

	// Check max length
	if max, ok := rules["max_items"].(int); ok {
		if value.Len() > max {
			return &ValidationError{
				Field:   "value",
				Message: fmt.Sprintf("must have at most %d items", max),
			}
		}
	}

	// Validate each item
	if itemRules, ok := rules["items"].(map[string]interface{}); ok {
		for i := 0; i < value.Len(); i++ {
			if err := Validate(value.Index(i).Interface(), itemRules); err != nil {
				return &ValidationError{
					Field:   fmt.Sprintf("items[%d]", i),
					Message: err.Error(),
				}
			}
		}
	}

	return nil
}

// validateMap validates a map value
func validateMap(value reflect.Value, rules map[string]interface{}) error {
	// Check min size
	if min, ok := rules["min_properties"].(int); ok {
		if value.Len() < min {
			return &ValidationError{
				Field:   "value",
				Message: fmt.Sprintf("must have at least %d properties", min),
			}
		}
	}

	// Check max size
	if max, ok := rules["max_properties"].(int); ok {
		if value.Len() > max {
			return &ValidationError{
				Field:   "value",
				Message: fmt.Sprintf("must have at most %d properties", max),
			}
		}
	}

	// Validate properties
	if properties, ok := rules["properties"].(map[string]interface{}); ok {
		iter := value.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			if propRules, ok := properties[key].(map[string]interface{}); ok {
				if err := Validate(iter.Value().Interface(), propRules); err != nil {
					return &ValidationError{
						Field:   fmt.Sprintf("properties.%s", key),
						Message: err.Error(),
					}
				}
			}
		}
	}

	return nil
}
