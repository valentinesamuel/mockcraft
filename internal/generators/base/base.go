package base

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/valentinesamuel/mockcraft/internal/generators/types"
)

// BaseGenerator provides basic functionality for generating fake data
type BaseGenerator struct {
	Faker *gofakeit.Faker
}

// NewBaseGenerator creates a new BaseGenerator instance
func NewBaseGenerator() *BaseGenerator {
	return &BaseGenerator{
		Faker: gofakeit.New(0),
	}
}

// GetAvailableTypes returns all available types that can be generated
func (g *BaseGenerator) GetAvailableTypes() []string {
	var typeList []string

	// Add gofakeit types
	t := reflect.TypeOf(g.Faker)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if method.Type.NumOut() > 0 {
			typeList = append(typeList, method.Name)
		}
	}

	// Add health types
	for t := range types.HealthTypes {
		typeList = append(typeList, t)
	}

	// Add custom types
	typeList = append(typeList,
		"password", "sentence", "shuffle_strings",
		"phone", "address", "ssn",
		"uuid", "mac_address", "domain",
		"word", "paragraph",
		"random_int", "random_float",
	)

	return typeList
}

// GenerateByType generates data of the specified type
func (g *BaseGenerator) GenerateByType(dataType string, params map[string]interface{}) (interface{}, error) {
	// Handle custom types
	switch dataType {
	case "password":
		length := 12 // default length
		if l, ok := params["length"].(int); ok {
			length = l
		}
		return g.generatePassword(length), nil

	case "sentence":
		wordCount := 6 // default word count
		if wc, ok := params["word_count"].(int); ok {
			wordCount = wc
		}
		return g.generateSentence(wordCount), nil

	case "shuffle_strings":
		strings, ok := params["strings"].(string)
		if !ok {
			return nil, fmt.Errorf("strings parameter is required for shuffle_strings")
		}
		return g.shuffleStrings(strings), nil

	case "phone":
		format := "international" // default format
		if f, ok := params["format"].(string); ok {
			format = f
		}
		return g.generatePhone(format), nil

	case "address":
		return g.generateAddress(), nil

	case "ssn":
		return g.generateSSN(), nil

	case "uuid":
		version := 4 // default version
		if v, ok := params["version"].(int); ok {
			version = v
		}
		return g.generateUUID(version), nil

	case "mac_address":
		return g.generateMACAddress(), nil

	case "domain":
		tld := "com" // default TLD
		if t, ok := params["tld"].(string); ok {
			tld = t
		}
		return g.generateDomain(tld), nil

	case "word":
		minLength := 4 // default min length
		maxLength := 8 // default max length
		if min, ok := params["min_length"].(int); ok {
			minLength = min
		}
		if max, ok := params["max_length"].(int); ok {
			maxLength = max
		}
		return g.generateWord(minLength, maxLength), nil

	case "paragraph":
		sentenceCount := 3 // default sentence count
		if sc, ok := params["sentence_count"].(int); ok {
			sentenceCount = sc
		}
		return g.generateParagraph(sentenceCount), nil

	case "random_int":
		min, ok1 := params["min"].(int)
		max, ok2 := params["max"].(int)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("min and max parameters are required for random_int")
		}
		return g.generateRandomInt(min, max), nil

	case "random_float":
		min, ok1 := params["min"].(float64)
		max, ok2 := params["max"].(float64)
		precision := 2 // default precision
		if p, ok := params["precision"].(int); ok {
			precision = p
		}
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("min and max parameters are required for random_float")
		}
		return g.generateRandomFloat(min, max, precision), nil
	}

	// Check if it's a health type
	if _, ok := types.HealthTypes[dataType]; ok {
		return g.generateHealthType(dataType)
	}

	// Convert snake_case to PascalCase for gofakeit methods
	dataType = g.convertToPascalCase(dataType)

	// Try to find and call the method
	method := reflect.ValueOf(g.Faker).MethodByName(dataType)
	if !method.IsValid() {
		return nil, fmt.Errorf("type '%s' not found", dataType)
	}

	// Call the method
	results := method.Call(nil)
	if len(results) == 0 {
		return nil, fmt.Errorf("method '%s' returned no values", dataType)
	}

	return results[0].Interface(), nil
}

// generatePassword generates a random password of specified length
func (g *BaseGenerator) generatePassword(length int) string {
	const (
		lowerChars = "abcdefghijklmnopqrstuvwxyz"
		upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numChars   = "0123456789"
		specChars  = "!@#$%^&*()_+-=[]{}|;:,.<>?"
		allChars   = lowerChars + upperChars + numChars + specChars
	)

	// Ensure at least one of each character type
	password := make([]byte, length)
	password[0] = lowerChars[g.Faker.IntRange(0, len(lowerChars)-1)]
	password[1] = upperChars[g.Faker.IntRange(0, len(upperChars)-1)]
	password[2] = numChars[g.Faker.IntRange(0, len(numChars)-1)]
	password[3] = specChars[g.Faker.IntRange(0, len(specChars)-1)]

	// Fill the rest with random characters
	for i := 4; i < length; i++ {
		password[i] = allChars[g.Faker.IntRange(0, len(allChars)-1)]
	}

	// Shuffle the password
	g.Faker.ShuffleAnySlice(password)

	return string(password)
}

// generateSentence generates a random sentence with specified word count
func (g *BaseGenerator) generateSentence(wordCount int) string {
	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		words[i] = g.Faker.Word()
	}
	return strings.Join(words, " ")
}

// shuffleStrings shuffles a comma-separated list of strings
func (g *BaseGenerator) shuffleStrings(input string) string {
	items := strings.Split(input, ",")
	g.Faker.ShuffleAnySlice(items)
	return strings.Join(items, ",")
}

// generatePhone generates a random phone number in the specified format
func (g *BaseGenerator) generatePhone(format string) string {
	switch format {
	case "international":
		return fmt.Sprintf("+%d (%d) %d-%d",
			g.Faker.IntRange(1, 99),
			g.Faker.IntRange(100, 999),
			g.Faker.IntRange(100, 999),
			g.Faker.IntRange(1000, 9999),
		)
	case "national":
		return fmt.Sprintf("(%d) %d-%d",
			g.Faker.IntRange(100, 999),
			g.Faker.IntRange(100, 999),
			g.Faker.IntRange(1000, 9999),
		)
	case "local":
		return fmt.Sprintf("%d-%d",
			g.Faker.IntRange(100, 999),
			g.Faker.IntRange(1000, 9999),
		)
	default:
		return g.generatePhone("international")
	}
}

// generateAddress generates a random address
func (g *BaseGenerator) generateAddress() string {
	return fmt.Sprintf("%d %s, %s, %s %s",
		g.Faker.IntRange(1, 9999),
		g.Faker.Street(),
		g.Faker.City(),
		g.Faker.State(),
		g.Faker.Zip(),
	)
}

// generateSSN generates a random Social Security Number
func (g *BaseGenerator) generateSSN() string {
	return fmt.Sprintf("%03d-%02d-%04d",
		g.Faker.IntRange(1, 999),
		g.Faker.IntRange(1, 99),
		g.Faker.IntRange(1, 9999),
	)
}

// generateUUID generates a random UUID of the specified version
func (g *BaseGenerator) generateUUID(version int) string {
	switch version {
	case 1:
		return g.Faker.UUID()
	case 3:
		return g.Faker.UUID()
	case 4:
		return g.Faker.UUID()
	case 5:
		return g.Faker.UUID()
	default:
		return g.Faker.UUID()
	}
}

// generateMACAddress generates a random MAC address
func (g *BaseGenerator) generateMACAddress() string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		g.Faker.IntRange(0, 255),
		g.Faker.IntRange(0, 255),
		g.Faker.IntRange(0, 255),
		g.Faker.IntRange(0, 255),
		g.Faker.IntRange(0, 255),
		g.Faker.IntRange(0, 255),
	)
}

// generateDomain generates a random domain name with the specified TLD
func (g *BaseGenerator) generateDomain(tld string) string {
	return fmt.Sprintf("%s.%s",
		g.Faker.DomainName(),
		tld,
	)
}

// generateWord generates a random word with length between min and max
func (g *BaseGenerator) generateWord(minLength, maxLength int) string {
	word := g.Faker.Word()
	for len(word) < minLength || len(word) > maxLength {
		word = g.Faker.Word()
	}
	return word
}

// generateParagraph generates a random paragraph with the specified number of sentences
func (g *BaseGenerator) generateParagraph(sentenceCount int) string {
	sentences := make([]string, sentenceCount)
	for i := 0; i < sentenceCount; i++ {
		sentences[i] = g.generateSentence(g.Faker.IntRange(5, 15))
	}
	return strings.Join(sentences, " ")
}

// generateRandomInt generates a random integer between min and max (inclusive)
func (g *BaseGenerator) generateRandomInt(min, max int) int {
	return g.Faker.IntRange(min, max)
}

// generateRandomFloat generates a random float between min and max (inclusive) with specified precision
func (g *BaseGenerator) generateRandomFloat(min, max float64, precision int) float64 {
	value := g.Faker.Float64Range(min, max)
	format := fmt.Sprintf("%%.%df", precision)
	result, _ := strconv.ParseFloat(fmt.Sprintf(format, value), 64)
	return result
}

// GenerateStruct generates values for a struct based on mock tags
func (g *BaseGenerator) GenerateStruct(v interface{}, params map[string]interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("input must be a pointer to a struct")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Get the mock tag
		mockTag := fieldType.Tag.Get("mock")
		if mockTag == "" {
			continue
		}

		// Parse tag for parameters
		tagParams := make(map[string]interface{})
		if mockTag != "" {
			parts := strings.Split(mockTag, ",")
			mockTag = parts[0] // First part is the type

			// Parse additional parameters from tag
			for _, part := range parts[1:] {
				if strings.HasPrefix(part, "length=") {
					if length, ok := params["length"].(int); ok {
						tagParams["length"] = length
					}
				} else if strings.HasPrefix(part, "word_count=") {
					if wordCount, ok := params["word_count"].(int); ok {
						tagParams["word_count"] = wordCount
					}
				} else if strings.HasPrefix(part, "strings=") {
					if strings, ok := params["strings"].(string); ok {
						tagParams["strings"] = strings
					}
				} else if strings.HasPrefix(part, "format=") {
					if format, ok := params["format"].(string); ok {
						tagParams["format"] = format
					}
				} else if strings.HasPrefix(part, "country=") {
					if country, ok := params["country"].(string); ok {
						tagParams["country"] = country
					}
				} else if strings.HasPrefix(part, "version=") {
					if version, ok := params["version"].(int); ok {
						tagParams["version"] = version
					}
				} else if strings.HasPrefix(part, "tld=") {
					if tld, ok := params["tld"].(string); ok {
						tagParams["tld"] = tld
					}
				} else if strings.HasPrefix(part, "min_length=") {
					if minLength, ok := params["min_length"].(int); ok {
						tagParams["min_length"] = minLength
					}
				} else if strings.HasPrefix(part, "max_length=") {
					if maxLength, ok := params["max_length"].(int); ok {
						tagParams["max_length"] = maxLength
					}
				} else if strings.HasPrefix(part, "sentence_count=") {
					if sentenceCount, ok := params["sentence_count"].(int); ok {
						tagParams["sentence_count"] = sentenceCount
					}
				} else if strings.HasPrefix(part, "min=") {
					if min, ok := params["min"].(float64); ok {
						tagParams["min"] = min
					} else if min, ok := params["min"].(int); ok {
						tagParams["min"] = float64(min)
					}
				} else if strings.HasPrefix(part, "max=") {
					if max, ok := params["max"].(float64); ok {
						tagParams["max"] = max
					} else if max, ok := params["max"].(int); ok {
						tagParams["max"] = float64(max)
					}
				} else if strings.HasPrefix(part, "precision=") {
					if precision, ok := params["precision"].(int); ok {
						tagParams["precision"] = precision
					}
				}
			}
		}

		// Generate value based on the mock tag and parameters
		if field.CanSet() {
			value, err := g.GenerateByType(mockTag, tagParams)
			if err != nil {
				return fmt.Errorf("error generating value for field %s: %v", fieldType.Name, err)
			}

			// Set the field value
			field.Set(reflect.ValueOf(value))
		}
	}

	return nil
}

// convertToPascalCase converts snake_case to PascalCase
func (g *BaseGenerator) convertToPascalCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, "")
}

// generateHealthType generates health-related data
func (g *BaseGenerator) generateHealthType(dataType string) (interface{}, error) {
	switch dataType {
	case "systolic":
		return g.Faker.IntRange(90, 140), nil
	case "diastolic":
		return g.Faker.IntRange(60, 90), nil
	case "blood_pressure_unit":
		return "mmHg", nil
	case "heart_rate":
		return g.Faker.IntRange(60, 100), nil
	case "heart_rate_unit":
		return "bpm", nil
	case "temperature":
		return g.Faker.Float64Range(97.0, 99.0), nil
	case "temperature_unit":
		return "°F", nil
	case "respiratory_rate":
		return g.Faker.IntRange(12, 20), nil
	case "respiratory_unit":
		return "breaths/min", nil
	default:
		return nil, fmt.Errorf("unknown health type: %s", dataType)
	}
}
