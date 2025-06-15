package generators

import (
	"fmt"
	"strconv"
	"strings"
)

// ParameterType represents the type of a parameter
type ParameterType string

const (
	ParamTypeString      ParameterType = "string"
	ParamTypeInt         ParameterType = "int"
	ParamTypeFloat       ParameterType = "float"
	ParamTypeBool        ParameterType = "bool"
	ParamTypeSelect      ParameterType = "select"      // For parameters with predefined options
	ParamTypeStringSlice ParameterType = "string_slice" // For array of strings
)

// ParameterDef defines a parameter for a generator
type ParameterDef struct {
	Name        string        `json:"name"`
	Type        ParameterType `json:"type"`
	Description string        `json:"description"`
	Required    bool          `json:"required"`
	Default     interface{}   `json:"default,omitempty"`
	Options     []string      `json:"options,omitempty"`     // For select type
	Min         interface{}   `json:"min,omitempty"`         // For int/float types
	Max         interface{}   `json:"max,omitempty"`         // For int/float types
	Example     interface{}   `json:"example,omitempty"`
}

// GeneratorInfo contains information about a generator and its parameters
type GeneratorInfo struct {
	Name        string         `json:"name"`
	Industry    string         `json:"industry"`
	Description string         `json:"description"`
	Example     string         `json:"example"`
	Parameters  []ParameterDef `json:"parameters"`
}

// ParameterRegistry holds parameter definitions for all generators
type ParameterRegistry struct {
	generators map[string]map[string]GeneratorInfo // industry -> generator -> info
}

// NewParameterRegistry creates a new parameter registry
func NewParameterRegistry() *ParameterRegistry {
	pr := &ParameterRegistry{
		generators: make(map[string]map[string]GeneratorInfo),
	}
	pr.registerAllParameters()
	return pr
}

// RegisterGenerator registers a generator with its parameter definitions
func (pr *ParameterRegistry) RegisterGenerator(industry, name string, info GeneratorInfo) {
	if pr.generators[industry] == nil {
		pr.generators[industry] = make(map[string]GeneratorInfo)
	}
	info.Name = name
	info.Industry = industry
	pr.generators[industry][name] = info
}

// GetGeneratorInfo returns parameter information for a specific generator
func (pr *ParameterRegistry) GetGeneratorInfo(industry, generator string) (*GeneratorInfo, error) {
	if industryGens, exists := pr.generators[industry]; exists {
		if info, exists := industryGens[generator]; exists {
			return &info, nil
		}
	}
	return nil, fmt.Errorf("generator %s not found in industry %s", generator, industry)
}

// GetAllGeneratorInfo returns all generator information
func (pr *ParameterRegistry) GetAllGeneratorInfo() map[string]map[string]GeneratorInfo {
	return pr.generators
}

// ValidateParameters validates parameters against the generator definition
func (pr *ParameterRegistry) ValidateParameters(industry, generator string, params map[string]interface{}) error {
	info, err := pr.GetGeneratorInfo(industry, generator)
	if err != nil {
		return err
	}

	// Check required parameters
	for _, paramDef := range info.Parameters {
		if paramDef.Required {
			if _, exists := params[paramDef.Name]; !exists {
				return fmt.Errorf("required parameter '%s' is missing", paramDef.Name)
			}
		}
	}

	// Validate parameter types and ranges
	for paramName, paramValue := range params {
		var paramDef *ParameterDef
		for i := range info.Parameters {
			if info.Parameters[i].Name == paramName {
				paramDef = &info.Parameters[i]
				break
			}
		}

		if paramDef == nil {
			continue // Unknown parameter, ignore for now
		}

		if err := pr.validateParameterValue(paramDef, paramValue); err != nil {
			return fmt.Errorf("invalid value for parameter '%s': %w", paramName, err)
		}
	}

	return nil
}

// validateParameterValue validates a single parameter value
func (pr *ParameterRegistry) validateParameterValue(paramDef *ParameterDef, value interface{}) error {
	switch paramDef.Type {
	case ParamTypeInt:
		if v, ok := value.(int); ok {
			if paramDef.Min != nil {
				if min, ok := paramDef.Min.(int); ok && v < min {
					return fmt.Errorf("value %d is less than minimum %d", v, min)
				}
			}
			if paramDef.Max != nil {
				if max, ok := paramDef.Max.(int); ok && v > max {
					return fmt.Errorf("value %d is greater than maximum %d", v, max)
				}
			}
		} else {
			return fmt.Errorf("expected integer, got %T", value)
		}
	case ParamTypeFloat:
		if v, ok := value.(float64); ok {
			if paramDef.Min != nil {
				if min, ok := paramDef.Min.(float64); ok && v < min {
					return fmt.Errorf("value %f is less than minimum %f", v, min)
				}
			}
			if paramDef.Max != nil {
				if max, ok := paramDef.Max.(float64); ok && v > max {
					return fmt.Errorf("value %f is greater than maximum %f", v, max)
				}
			}
		} else {
			return fmt.Errorf("expected float, got %T", value)
		}
	case ParamTypeSelect:
		if v, ok := value.(string); ok {
			for _, option := range paramDef.Options {
				if v == option {
					return nil
				}
			}
			return fmt.Errorf("value '%s' is not one of the allowed options: %v", v, paramDef.Options)
		} else {
			return fmt.Errorf("expected string for select parameter, got %T", value)
		}
	case ParamTypeStringSlice:
		if _, ok := value.([]string); !ok {
			return fmt.Errorf("expected string slice, got %T", value)
		}
	}
	return nil
}

// ConvertAndValidateParams converts CLI flags to proper parameter types and validates them
func (pr *ParameterRegistry) ConvertAndValidateParams(industry, generator string, rawParams map[string]interface{}) (map[string]interface{}, error) {
	info, err := pr.GetGeneratorInfo(industry, generator)
	if err != nil {
		return rawParams, nil // If no parameter info, return as-is
	}

	convertedParams := make(map[string]interface{})
	
	// Set defaults first
	for _, paramDef := range info.Parameters {
		if paramDef.Default != nil {
			convertedParams[paramDef.Name] = paramDef.Default
		}
	}

	// Convert and validate provided parameters
	for paramName, rawValue := range rawParams {
		var paramDef *ParameterDef
		for i := range info.Parameters {
			if info.Parameters[i].Name == paramName {
				paramDef = &info.Parameters[i]
				break
			}
		}

		if paramDef == nil {
			convertedParams[paramName] = rawValue // Unknown parameter, keep as-is
			continue
		}

		convertedValue, err := pr.convertParameterValue(paramDef, rawValue)
		if err != nil {
			return nil, fmt.Errorf("failed to convert parameter '%s': %w", paramName, err)
		}

		convertedParams[paramName] = convertedValue
	}

	// Validate the final parameters
	if err := pr.ValidateParameters(industry, generator, convertedParams); err != nil {
		return nil, err
	}

	return convertedParams, nil
}

// convertParameterValue converts a raw parameter value to the expected type
func (pr *ParameterRegistry) convertParameterValue(paramDef *ParameterDef, rawValue interface{}) (interface{}, error) {
	// If already the right type, return as-is
	switch paramDef.Type {
	case ParamTypeInt:
		switch v := rawValue.(type) {
		case int:
			return v, nil
		case string:
			if v == "" {
				return 0, nil
			}
			return strconv.Atoi(v)
		case float64:
			return int(v), nil
		}
	case ParamTypeFloat:
		switch v := rawValue.(type) {
		case float64:
			return v, nil
		case string:
			if v == "" {
				return 0.0, nil
			}
			return strconv.ParseFloat(v, 64)
		case int:
			return float64(v), nil
		}
	case ParamTypeBool:
		switch v := rawValue.(type) {
		case bool:
			return v, nil
		case string:
			return strconv.ParseBool(v)
		}
	case ParamTypeString, ParamTypeSelect:
		return fmt.Sprintf("%v", rawValue), nil
	case ParamTypeStringSlice:
		switch v := rawValue.(type) {
		case []string:
			return v, nil
		case string:
			// Parse comma-separated string
			if v == "" {
				return []string{}, nil
			}
			parts := strings.Split(v, ",")
			result := make([]string, len(parts))
			for i, part := range parts {
				result[i] = strings.TrimSpace(part)
			}
			return result, nil
		case []interface{}:
			// Convert interface slice to string slice
			result := make([]string, len(v))
			for i, item := range v {
				result[i] = fmt.Sprintf("%v", item)
			}
			return result, nil
		}
	}

	return rawValue, nil
}

// ApplyTextTransformations applies text transformations like uppercase, lowercase, etc.
func ApplyTextTransformations(value interface{}, params map[string]interface{}) interface{} {
	str, ok := value.(string)
	if !ok {
		return value
	}

	// Apply text transformations
	if uppercase, ok := params["uppercase"].(bool); ok && uppercase {
		str = strings.ToUpper(str)
	} else if lowercase, ok := params["lowercase"].(bool); ok && lowercase {
		str = strings.ToLower(str)
	} else if capitalize, ok := params["capitalize"].(bool); ok && capitalize {
		str = strings.Title(str)
	}

	// Apply prefix and suffix
	if prefix, ok := params["prefix"].(string); ok && prefix != "" {
		str = prefix + str
	}
	if suffix, ok := params["suffix"].(string); ok && suffix != "" {
		str = str + suffix
	}

	return str
}

// Global parameter registry instance
var globalParamRegistry *ParameterRegistry

// GetGlobalParameterRegistry returns the global parameter registry
func GetGlobalParameterRegistry() *ParameterRegistry {
	if globalParamRegistry == nil {
		globalParamRegistry = NewParameterRegistry()
	}
	return globalParamRegistry
}