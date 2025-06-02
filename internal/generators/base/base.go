package base

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/valentinesamuel/mockcraft/internal/generators/health"
	"github.com/valentinesamuel/mockcraft/internal/generators/types"
	"github.com/valentinesamuel/mockcraft/internal/interfaces"
)

// BaseGenerator implements the Generator interface using gofakeit
type BaseGenerator struct {
	faker *gofakeit.Faker
}

// NewBaseGenerator creates a new BaseGenerator instance
func NewBaseGenerator() interfaces.Generator {
	return &BaseGenerator{
		faker: gofakeit.New(0),
	}
}

// GetAvailableTypes returns all available types for this generator
func (g *BaseGenerator) GetAvailableTypes() []string {
	return []string{
		"first_name",
		"last_name",
		"email",
		"phone",
		"address",
		"city",
		"state",
		"zip",
		"country",
		"company",
		"job_title",
		"domain",
		"url",
		"ip",
		"mac",
		"uuid",
		"password",
		"word",
		"sentence",
		"paragraph",
		"date",
		"time",
		"datetime",
		"random_int",
		"random_float",
		"boolean",
		"color",
		"credit_card",
		"ssn",
		"ein",
	}
}

// applyTextTransformations applies text transformations to the generated data
func (g *BaseGenerator) applyTextTransformations(data interface{}, params map[string]interface{}) interface{} {
	// Convert to string if possible
	var text string
	switch v := data.(type) {
	case string:
		text = v
	case fmt.Stringer:
		text = v.String()
	default:
		return data // Return original if not a string
	}

	// Apply transformations
	if params["uppercase"] == true {
		text = strings.ToUpper(text)
	} else if params["lowercase"] == true {
		text = strings.ToLower(text)
	} else if params["capitalize"] == true {
		text = strings.Title(strings.ToLower(text))
	}

	// Apply prefix and suffix
	if prefix, ok := params["prefix"].(string); ok && prefix != "" {
		text = prefix + text
	}
	if suffix, ok := params["suffix"].(string); ok && suffix != "" {
		text = text + suffix
	}

	return text
}

// GenerateByType generates data of the specified type
func (g *BaseGenerator) GenerateByType(dataType string, params map[string]interface{}) (interface{}, error) {
	// Validate parameters
	if err := g.validateParameters(dataType, params); err != nil {
		return nil, err
	}

	var result interface{}
	var err error

	// Handle custom types
	switch dataType {
	case "password":
		length := 12 // default length
		if l, ok := params["length"].(int); ok {
			length = l
		}
		result = g.generatePassword(length)

	case "sentence":
		wordCount := 6 // default word count
		if wc, ok := params["word_count"].(int); ok {
			wordCount = wc
		}
		result = g.generateSentence(wordCount)

	case "shuffle_strings":
		strings, ok := params["strings"].(string)
		if !ok {
			return nil, fmt.Errorf("strings parameter is required for shuffle_strings")
		}
		result = g.shuffleStrings(strings)

	case "phone":
		format := "international" // default format
		if f, ok := params["format"].(string); ok {
			format = f
		}
		result = g.generatePhone(format)

	case "address":
		result = g.generateAddress()

	case "ssn":
		result = g.generateSSN()

	case "uuid":
		version := 4 // default version
		if v, ok := params["version"].(int); ok {
			version = v
		}
		result = g.generateUUID(version)

	case "mac_address":
		result = g.generateMACAddress()

	case "domain":
		tld := "com" // default TLD
		if t, ok := params["tld"].(string); ok {
			tld = t
		}
		result = g.generateDomain(tld)

	case "word":
		minLength := 4 // default min length
		maxLength := 8 // default max length
		if min, ok := params["min_length"].(int); ok {
			minLength = min
		}
		if max, ok := params["max_length"].(int); ok {
			maxLength = max
		}
		result = g.generateWord(minLength, maxLength)

	case "paragraph":
		sentenceCount := 3 // default sentence count
		if sc, ok := params["sentence_count"].(int); ok {
			sentenceCount = sc
		}
		result = g.generateParagraph(sentenceCount)

	case "random_int":
		min, ok1 := params["min"].(int)
		max, ok2 := params["max"].(int)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("min and max parameters are required for random_int")
		}
		result = g.generateRandomInt(min, max)

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
		result = g.generateRandomFloat(min, max, precision)

	default:
		// Check if it's a health type
		if _, ok := types.HealthTypes[dataType]; ok {
			result, err = g.generateHealthType(dataType)
			if err != nil {
				return nil, err
			}
		} else {
			// Convert snake_case to PascalCase for gofakeit methods
			dataType = g.convertToPascalCase(dataType)

			// Try to find and call the method
			method := reflect.ValueOf(g.faker).MethodByName(dataType)
			if !method.IsValid() {
				return nil, fmt.Errorf("type '%s' not found", dataType)
			}

			// Call the method
			results := method.Call(nil)
			if len(results) == 0 {
				return nil, fmt.Errorf("method '%s' returned no values", dataType)
			}

			result = results[0].Interface()
		}
	}

	// Apply text transformations if applicable
	return g.applyTextTransformations(result, params), nil
}

// validateParameters validates the parameters for a given type
func (g *BaseGenerator) validateParameters(dataType string, params map[string]interface{}) error {
	// Get type definition
	typeDef := types.GetTypeByName(dataType)
	if typeDef == nil {
		// If no type definition exists, assume it's a gofakeit type
		return nil
	}

	// Check required parameters
	for _, param := range typeDef.Parameters {
		if param.Required {
			if _, ok := params[param.Name]; !ok {
				return fmt.Errorf("required parameter '%s' is missing for type '%s'", param.Name, dataType)
			}
		}
	}

	// Validate parameter types
	for name, value := range params {
		// Find parameter definition
		var paramDef *types.Parameter
		for _, p := range typeDef.Parameters {
			if p.Name == name {
				paramDef = &p
				break
			}
		}

		if paramDef == nil {
			return fmt.Errorf("unknown parameter '%s' for type '%s'", name, dataType)
		}

		// Validate parameter type
		if err := g.validateParameterType(name, value, paramDef.Type); err != nil {
			return err
		}
	}

	return nil
}

// validateParameterType validates that a parameter value matches its expected type
func (g *BaseGenerator) validateParameterType(name string, value interface{}, expectedType string) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter '%s' must be a string", name)
		}
	case "int":
		if _, ok := value.(int); !ok {
			return fmt.Errorf("parameter '%s' must be an integer", name)
		}
	case "float":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("parameter '%s' must be a float", name)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("parameter '%s' must be a boolean", name)
		}
	default:
		return fmt.Errorf("unknown parameter type '%s' for parameter '%s'", expectedType, name)
	}
	return nil
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
	password[0] = lowerChars[g.faker.IntRange(0, len(lowerChars)-1)]
	password[1] = upperChars[g.faker.IntRange(0, len(upperChars)-1)]
	password[2] = numChars[g.faker.IntRange(0, len(numChars)-1)]
	password[3] = specChars[g.faker.IntRange(0, len(specChars)-1)]

	// Fill the rest with random characters
	for i := 4; i < length; i++ {
		password[i] = allChars[g.faker.IntRange(0, len(allChars)-1)]
	}

	// Shuffle the password
	g.faker.ShuffleAnySlice(password)

	return string(password)
}

// generateSentence generates a random sentence with specified word count
func (g *BaseGenerator) generateSentence(wordCount int) string {
	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		words[i] = g.faker.Word()
	}
	return strings.Join(words, " ")
}

// shuffleStrings shuffles a comma-separated list of strings
func (g *BaseGenerator) shuffleStrings(input string) string {
	items := strings.Split(input, ",")
	g.faker.ShuffleAnySlice(items)
	return strings.Join(items, ",")
}

// generatePhone generates a random phone number in the specified format
func (g *BaseGenerator) generatePhone(format string) string {
	switch format {
	case "international":
		return fmt.Sprintf("+%d (%d) %d-%d",
			g.faker.IntRange(1, 99),
			g.faker.IntRange(100, 999),
			g.faker.IntRange(100, 999),
			g.faker.IntRange(1000, 9999),
		)
	case "national":
		return fmt.Sprintf("(%d) %d-%d",
			g.faker.IntRange(100, 999),
			g.faker.IntRange(100, 999),
			g.faker.IntRange(1000, 9999),
		)
	case "local":
		return fmt.Sprintf("%d-%d",
			g.faker.IntRange(100, 999),
			g.faker.IntRange(1000, 9999),
		)
	default:
		return g.generatePhone("international")
	}
}

// generateAddress generates a random address
func (g *BaseGenerator) generateAddress() string {
	return fmt.Sprintf("%d %s, %s, %s %s",
		g.faker.IntRange(1, 9999),
		g.faker.Street(),
		g.faker.City(),
		g.faker.State(),
		g.faker.Zip(),
	)
}

// generateSSN generates a random Social Security Number
func (g *BaseGenerator) generateSSN() string {
	return fmt.Sprintf("%03d-%02d-%04d",
		g.faker.IntRange(1, 999),
		g.faker.IntRange(1, 99),
		g.faker.IntRange(1, 9999),
	)
}

// generateUUID generates a random UUID of the specified version
func (g *BaseGenerator) generateUUID(version int) string {
	switch version {
	case 1:
		return g.faker.UUID()
	case 3:
		return g.faker.UUID()
	case 4:
		return g.faker.UUID()
	case 5:
		return g.faker.UUID()
	default:
		return g.faker.UUID()
	}
}

// generateMACAddress generates a random MAC address
func (g *BaseGenerator) generateMACAddress() string {
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",
		g.faker.IntRange(0, 255),
		g.faker.IntRange(0, 255),
		g.faker.IntRange(0, 255),
		g.faker.IntRange(0, 255),
		g.faker.IntRange(0, 255),
		g.faker.IntRange(0, 255),
	)
}

// generateDomain generates a random domain name with the specified TLD
func (g *BaseGenerator) generateDomain(tld string) string {
	return fmt.Sprintf("%s.%s",
		g.faker.DomainName(),
		tld,
	)
}

// generateWord generates a random word with length between min and max
func (g *BaseGenerator) generateWord(minLength, maxLength int) string {
	word := g.faker.Word()
	for len(word) < minLength || len(word) > maxLength {
		word = g.faker.Word()
	}
	return word
}

// generateParagraph generates a random paragraph with the specified number of sentences
func (g *BaseGenerator) generateParagraph(sentenceCount int) string {
	sentences := make([]string, sentenceCount)
	for i := 0; i < sentenceCount; i++ {
		sentences[i] = g.generateSentence(g.faker.IntRange(5, 15))
	}
	return strings.Join(sentences, " ")
}

// generateRandomInt generates a random integer between min and max (inclusive)
func (g *BaseGenerator) generateRandomInt(min, max int) int {
	return g.faker.IntRange(min, max)
}

// generateRandomFloat generates a random float between min and max (inclusive) with specified precision
func (g *BaseGenerator) generateRandomFloat(min, max float64, precision int) float64 {
	value := g.faker.Float64Range(min, max)
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
	healthGen := health.NewHealthGenerator(g)

	switch dataType {
	case "blood_type":
		return healthGen.GenerateBloodType(), nil
	case "medical_condition":
		return healthGen.GenerateMedicalCondition(), nil
	case "medication":
		return healthGen.GenerateMedication(), nil
	case "symptom":
		return healthGen.GenerateSymptom(), nil
	case "diagnosis":
		return healthGen.GenerateDiagnosis(), nil
	case "allergy":
		return healthGen.GenerateAllergy(), nil
	case "lab_result":
		return healthGen.GenerateLabResult(), nil
	case "vital_sign":
		return healthGen.GenerateVitalSigns(), nil
	case "medical_record":
		return healthGen.GenerateMedicalRecord(), nil
	default:
		return nil, fmt.Errorf("unknown health type: %s", dataType)
	}
}

// SetSeed sets the random seed for reproducible results
func (g *BaseGenerator) SetSeed(seed int64) {
	g.faker = gofakeit.New(seed)
}
