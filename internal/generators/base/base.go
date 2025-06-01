package base

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/brianvoe/gofakeit/v7"
)

// BaseGenerator provides common functionality for all generators
type BaseGenerator struct {
	faker *gofakeit.Faker
}

// NewBaseGenerator creates a new base generator instance
func NewBaseGenerator() *BaseGenerator {
	return &BaseGenerator{
		faker: gofakeit.New(0),
	}
}

// GetFaker returns the faker instance
func (bg *BaseGenerator) GetFaker() *gofakeit.Faker {
	return bg.faker
}

// toPascalCase converts a snake_case string to PascalCase
func toPascalCase(s string) string {
	if len(s) == 0 {
		return s
	}
	// Split by underscore and capitalize each word
	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = string(unicode.ToUpper(rune(word[0]))) + word[1:]
		}
	}
	return strings.Join(words, "")
}

// GenerateByType generates data based on type string and parameters
func (bg *BaseGenerator) GenerateByType(dataType string, params map[string]interface{}) (interface{}, error) {
	// Use custom implementations for certain types
	switch dataType {
	case "password":
		return CustomPassword(params), nil
	case "sentence":
		return CustomSentence(params), nil
	case "shuffle_strings":
		return CustomShuffleStrings(params), nil
	}

	// Convert dataType to gofakeit function name
	funcName := toPascalCase(dataType)
	fmt.Printf("Looking for method: %s\n", funcName)

	// Get the function from gofakeit
	faker := bg.GetFaker()
	funcValue := reflect.ValueOf(faker).MethodByName(funcName)

	if !funcValue.IsValid() {
		// List available methods for debugging
		// t := reflect.TypeOf(faker)
		// fmt.Println("\nAvailable methods:")
		// for i := 0; i < t.NumMethod(); i++ {
		// 	method := t.Method(i)
		// 	if method.Type.NumOut() == 1 && !strings.HasPrefix(method.Name, "Get") {
		// 		fmt.Printf("- %s\n", method.Name)
		// 	}
		// }
		return nil, fmt.Errorf("type '%s' not found", dataType)
	}

	// Prepare arguments for the function call
	var args []reflect.Value
	if len(params) > 0 {
		// Convert params to function arguments
		funcType := funcValue.Type()
		for i := 0; i < funcType.NumIn(); i++ {
			paramType := funcType.In(i)
			if val, ok := params[fmt.Sprintf("arg%d", i)]; ok {
				args = append(args, reflect.ValueOf(val))
			} else {
				// Use zero value for missing parameters
				args = append(args, reflect.Zero(paramType))
			}
		}
	}

	// Call the function with arguments
	results := funcValue.Call(args)
	if len(results) == 0 {
		return nil, fmt.Errorf("function returned no values")
	}

	return results[0].Interface(), nil
}

// GetAvailableTypes returns all supported types
func (bg *BaseGenerator) GetAvailableTypes() []string {
	faker := bg.GetFaker()
	t := reflect.TypeOf(faker)

	var types []string
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		// Only include methods that return a single value and don't start with "Get"
		if method.Type.NumOut() == 1 && !strings.HasPrefix(method.Name, "Get") {
			// Convert PascalCase to lowercase
			var words []string
			word := ""
			for _, r := range method.Name {
				if unicode.IsUpper(r) && word != "" {
					words = append(words, strings.ToLower(word))
					word = string(unicode.ToLower(r))
				} else {
					word += string(unicode.ToLower(r))
				}
			}
			if word != "" {
				words = append(words, strings.ToLower(word))
			}
			types = append(types, strings.Join(words, ""))
		}
	}

	return types
}
