package base

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	// "github.com/valentinesamuel/mockcraft/internal/generators/industries/health"
	"github.com/valentinesamuel/mockcraft/internal/generators/interfaces"
	genregistry "github.com/valentinesamuel/mockcraft/internal/generators/registry"
	"github.com/valentinesamuel/mockcraft/internal/registry"
)

var globalRegistry = genregistry.NewIndustryRegistry()

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

// Auto-register when package is imported
func init() {
	registry.Register("base", NewBaseGenerator)

	// Register base generators
	globalRegistry.RegisterGenerator("base", "uuid", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.UUID(), nil
	})
	globalRegistry.RegisterGenerator("base", "firstname", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.FirstName(), nil
	})
	globalRegistry.RegisterGenerator("base", "lastname", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.LastName(), nil
	})
	globalRegistry.RegisterGenerator("base", "email", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Email(), nil
	})
	globalRegistry.RegisterGenerator("base", "phone", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Phone(), nil
	})
	globalRegistry.RegisterGenerator("base", "address", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Address().Address, nil
	})
	globalRegistry.RegisterGenerator("base", "company", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Company(), nil
	})
	globalRegistry.RegisterGenerator("base", "job_title", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.JobTitle(), nil
	})
	globalRegistry.RegisterGenerator("base", "date", func(params map[string]interface{}) (interface{}, error) {
		start := time.Now().AddDate(-1, 0, 0)
		end := time.Now()
		return gofakeit.DateRange(start, end).Format("2006-01-02"), nil
	})
	globalRegistry.RegisterGenerator("base", "datetime", func(params map[string]interface{}) (interface{}, error) {
		start := time.Now().AddDate(-1, 0, 0)
		end := time.Now()
		return gofakeit.DateRange(start, end).Format(time.RFC3339), nil
	})
	globalRegistry.RegisterGenerator("base", "time", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Date().Format("15:04:05"), nil
	})
	globalRegistry.RegisterGenerator("base", "timestamp", func(params map[string]interface{}) (interface{}, error) {
		start := time.Now().AddDate(-1, 0, 0)
		end := time.Now()
		return gofakeit.DateRange(start, end), nil
	})
	globalRegistry.RegisterGenerator("base", "number", func(params map[string]interface{}) (interface{}, error) {
		min := 0
		max := 100
		if minVal, ok := params["min"].(int); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(int); ok {
			max = maxVal
		}
		return rand.Intn(max-min+1) + min, nil
	})
	globalRegistry.RegisterGenerator("base", "float", func(params map[string]interface{}) (interface{}, error) {
		min := 0.0
		max := 100.0
		if minVal, ok := params["min"].(float64); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(float64); ok {
			max = maxVal
		}
		return min + rand.Float64()*(max-min), nil
	})
	globalRegistry.RegisterGenerator("base", "boolean", func(params map[string]interface{}) (interface{}, error) {
		return rand.Float32() < 0.5, nil
	})
	globalRegistry.RegisterGenerator("base", "text", func(params map[string]interface{}) (interface{}, error) {
		min := 10
		max := 100
		if minVal, ok := params["min"].(int); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(int); ok {
			max = maxVal
		}
		length := min + rand.Intn(max-min+1)
		chars := make([]string, length)
		for i := 0; i < length; i++ {
			chars[i] = gofakeit.RandomString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})
		}
		return strings.Join(chars, ""), nil
	})
	globalRegistry.RegisterGenerator("base", "paragraph", func(params map[string]interface{}) (interface{}, error) {
		count := 1
		if countVal, ok := params["count"].(int); ok {
			count = countVal
		}
		return gofakeit.Paragraph(count, count, 10, "\n"), nil
	})
	globalRegistry.RegisterGenerator("base", "sentence", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Sentence(10), nil
	})
	globalRegistry.RegisterGenerator("base", "word", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Word(), nil
	})
	globalRegistry.RegisterGenerator("base", "char", func(params map[string]interface{}) (interface{}, error) {
		return string(gofakeit.RandomString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})[0]), nil
	})
	globalRegistry.RegisterGenerator("base", "url", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.URL(), nil
	})
	globalRegistry.RegisterGenerator("base", "ip", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.IPv4Address(), nil
	})
	globalRegistry.RegisterGenerator("base", "domain", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.DomainName(), nil
	})
	globalRegistry.RegisterGenerator("base", "username", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Username(), nil
	})
	globalRegistry.RegisterGenerator("base", "foreign", func(params map[string]interface{}) (interface{}, error) {
		table, ok := params["table"].(string)
		if !ok {
			return nil, fmt.Errorf("missing required parameter 'table'")
		}

		column, ok := params["column"].(string)
		if !ok {
			return nil, fmt.Errorf("missing required parameter 'column'")
		}

		values := getForeignValues(table, column)
		if len(values) == 0 {
			return nil, fmt.Errorf("no values found for foreign key %s.%s", table, column)
		}

		// Use the first value for now - in a real implementation, you'd want to randomly select one
		return values[0], nil
	})
}

// getForeignValues returns a list of values for a foreign key relationship
func getForeignValues(table, column string) []interface{} {
	// TODO: Implement actual foreign key value lookup
	// For now, return some dummy values based on the column type
	switch column {
	case "id":
		return []interface{}{1, 2, 3, 4, 5}
	case "name":
		return []interface{}{"John", "Jane", "Bob", "Alice", "Charlie"}
	case "email":
		return []interface{}{"john@example.com", "jane@example.com", "bob@example.com", "alice@example.com", "charlie@example.com"}
	case "code":
		return []interface{}{"A", "B", "C", "D", "E"}
	default:
		return []interface{}{"value1", "value2", "value3", "value4", "value5"}
	}
}

// GetAvailableTypes returns the list of available generator types
func (g *BaseGenerator) GetAvailableTypes() []string {
	return []string{
		"uuid",
		"word",
		"sentence",
		"timestamp",
		"datetime",
		"date",
		"time",
		"number",
		"float",
		"boolean",
		"text",
		"paragraph",
		"foreign",
	}
}

// GenerateByType generates data of the specified type
func (g *BaseGenerator) GenerateByType(dataType string, params map[string]interface{}) (interface{}, error) {
	switch dataType {
	case "uuid":
		return g.faker.UUID(), nil

	case "word":
		return g.faker.Word(), nil

	case "sentence":
		return g.faker.Sentence(10), nil

	case "paragraph":
		count := 1
		if val, ok := params["count"].(int); ok {
			count = val
		}
		return g.faker.Paragraph(count, count, 10, "\n"), nil

	case "date":
		start := time.Now().AddDate(-1, 0, 0)
		end := time.Now()
		if startStr, ok := params["start"].(string); ok {
			if t, err := time.Parse("2006-01-02", startStr); err == nil {
				start = t
			}
		}
		if endStr, ok := params["end"].(string); ok {
			if t, err := time.Parse("2006-01-02", endStr); err == nil {
				end = t
			}
		}
		return g.faker.DateRange(start, end).Format("2006-01-02"), nil

	case "datetime", "timestamp":
		start := time.Now().AddDate(-1, 0, 0)
		end := time.Now()
		if startStr, ok := params["start"].(string); ok {
			if t, err := time.Parse(time.RFC3339, startStr); err == nil {
				start = t
			}
		}
		if endStr, ok := params["end"].(string); ok {
			if t, err := time.Parse(time.RFC3339, endStr); err == nil {
				end = t
			}
		}
		if dataType == "datetime" {
			return g.faker.DateRange(start, end).Format(time.RFC3339), nil
		}
		return g.faker.DateRange(start, end).Unix(), nil

	case "time":
		return g.faker.Date().Format("15:04:05"), nil

	case "number":
		min := 0
		max := 100
		if minVal, ok := params["min"].(int); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(int); ok {
			max = maxVal
		}
		return rand.Intn(max-min+1) + min, nil

	case "float":
		min := 0.0
		max := 100.0
		if minVal, ok := params["min"].(float64); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(float64); ok {
			max = maxVal
		}
		return min + rand.Float64()*(max-min), nil

	case "boolean":
		return rand.Float32() < 0.5, nil

	case "text":
		min := 10
		max := 100
		if minVal, ok := params["min"].(int); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(int); ok {
			max = maxVal
		}
		length := min + rand.Intn(max-min+1)
		chars := make([]string, length)
		for i := 0; i < length; i++ {
			chars[i] = gofakeit.RandomString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})
		}
		return strings.Join(chars, ""), nil

	case "foreign":
		table, ok := params["table"].(string)
		if !ok {
			return nil, fmt.Errorf("missing required parameter 'table'")
		}

		column, ok := params["column"].(string)
		if !ok {
			return nil, fmt.Errorf("missing required parameter 'column'")
		}

		values := getForeignValues(table, column)
		if len(values) == 0 {
			return nil, fmt.Errorf("no values found for foreign key %s.%s", table, column)
		}

		// Use the first value for now - in a real implementation, you'd want to randomly select one
		return values[0], nil

	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}
}

// validateParameters validates the parameters for a given data type
func (g *BaseGenerator) validateParameters(dataType string, params map[string]interface{}) error {
	switch dataType {
	case "number", "float":
		if min, ok := params["min"]; ok {
			if max, ok := params["max"]; ok {
				if reflect.TypeOf(min) != reflect.TypeOf(max) {
					return fmt.Errorf("min and max must be of the same type")
				}
				switch min.(type) {
				case int:
					if min.(int) > max.(int) {
						return fmt.Errorf("min must be less than or equal to max")
					}
				case float64:
					if min.(float64) > max.(float64) {
						return fmt.Errorf("min must be less than or equal to max")
					}
				default:
					return fmt.Errorf("min and max must be numbers")
				}
			}
		}
	case "text":
		if min, ok := params["min"].(int); ok {
			if max, ok := params["max"].(int); ok {
				if min > max {
					return fmt.Errorf("min must be less than or equal to max")
				}
			}
		}
	case "date", "datetime", "timestamp":
		if start, ok := params["start"].(string); ok {
			if end, ok := params["end"].(string); ok {
				var startTime, endTime time.Time
				var err error
				if dataType == "date" {
					startTime, err = time.Parse("2006-01-02", start)
					if err != nil {
						return fmt.Errorf("invalid start date format: %v", err)
					}
					endTime, err = time.Parse("2006-01-02", end)
					if err != nil {
						return fmt.Errorf("invalid end date format: %v", err)
					}
				} else {
					startTime, err = time.Parse(time.RFC3339, start)
					if err != nil {
						return fmt.Errorf("invalid start datetime format: %v", err)
					}
					endTime, err = time.Parse(time.RFC3339, end)
					if err != nil {
						return fmt.Errorf("invalid end datetime format: %v", err)
					}
				}
				if startTime.After(endTime) {
					return fmt.Errorf("start time must be before end time")
				}
			}
		}
	case "foreign":
		if _, ok := params["table"]; !ok {
			return fmt.Errorf("missing required parameter 'table'")
		}
		if _, ok := params["column"]; !ok {
			return fmt.Errorf("missing required parameter 'column'")
		}
	}
	return nil
}

// generateUUID generates a UUID of the specified version
func (g *BaseGenerator) generateUUID(version int) string {
	switch version {
	case 1:
		return g.faker.UUID()
	case 4:
		return g.faker.UUID()
	default:
		return g.faker.UUID()
	}
}

// generatePhone generates a phone number in the specified format
func (g *BaseGenerator) generatePhone(format string) string {
	switch format {
	case "e164":
		return "+1" + g.faker.Phone()
	case "national":
		return g.faker.Phone()
	default:
		return g.faker.Phone()
	}
}

// GenerateStruct generates data for a struct based on its tags
func (g *BaseGenerator) GenerateStruct(v interface{}, params map[string]interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("v must be a pointer to a struct")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get the generator type from the tag
		generatorType := fieldType.Tag.Get("generator")
		if generatorType == "" {
			continue
		}

		// Get the generator parameters from the tag
		generatorParams := make(map[string]interface{})
		if paramsStr := fieldType.Tag.Get("params"); paramsStr != "" {
			// Parse the params string into a map
			// Format: "key1=value1,key2=value2"
			pairs := strings.Split(paramsStr, ",")
			for _, pair := range pairs {
				kv := strings.Split(pair, "=")
				if len(kv) == 2 {
					generatorParams[kv[0]] = kv[1]
				}
			}
		}

		// Generate the value
		value, err := g.GenerateByType(generatorType, generatorParams)
		if err != nil {
			return fmt.Errorf("failed to generate value for field %s: %v", fieldType.Name, err)
		}

		// Set the field value
		if !field.CanSet() {
			continue
		}

		// Convert the generated value to the field type
		fieldValue := reflect.ValueOf(value)
		if fieldValue.Type().ConvertibleTo(field.Type()) {
			field.Set(fieldValue.Convert(field.Type()))
		} else {
			return fmt.Errorf("cannot convert generated value to field type for field %s", fieldType.Name)
		}
	}

	return nil
}

// SetSeed sets the random seed for the generator
func (g *BaseGenerator) SetSeed(seed int64) {
	g.faker = gofakeit.New(seed)
}

// generateBankName generates a random bank name
func (g *BaseGenerator) generateBankName() string {
	return g.faker.Company() + " Bank"
}

// generateAccountNumber generates a random account number
func (g *BaseGenerator) generateAccountNumber() string {
	return g.faker.RandomString([]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"})
}

// generateRoutingNumber generates a random routing number
func (g *BaseGenerator) generateRoutingNumber() string {
	return g.faker.RandomString([]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"})
}

// generateCreditCard generates a random credit card number
func (g *BaseGenerator) generateCreditCard() string {
	return g.faker.CreditCardNumber(nil)
}

// generateCurrency generates a random currency name
func (g *BaseGenerator) generateCurrency() string {
	currencies := []string{
		"US Dollar",
		"Euro",
		"British Pound",
		"Japanese Yen",
		"Australian Dollar",
		"Canadian Dollar",
		"Swiss Franc",
		"Chinese Yuan",
		"Indian Rupee",
		"Brazilian Real",
	}
	return g.faker.RandomString(currencies)
}

// generateCurrencyCode generates a random currency code
func (g *BaseGenerator) generateCurrencyCode() string {
	codes := []string{"USD", "EUR", "GBP", "JPY", "AUD", "CAD", "CHF", "CNY", "INR", "BRL"}
	return g.faker.RandomString(codes)
}

// generateStockSymbol generates a random stock symbol
func (g *BaseGenerator) generateStockSymbol() string {
	symbols := []string{"AAPL", "GOOGL", "MSFT", "AMZN", "FB", "TSLA", "NVDA", "PYPL", "INTC", "AMD"}
	return g.faker.RandomString(symbols)
}

// generateStockPrice generates a random stock price
func (g *BaseGenerator) generateStockPrice() float64 {
	return g.faker.Float64Range(10.0, 1000.0)
}

// generateTransactionType generates a random transaction type
func (g *BaseGenerator) generateTransactionType() string {
	types := []string{
		"deposit",
		"withdrawal",
		"transfer",
		"payment",
		"refund",
	}
	return g.faker.RandomString(types)
}

// generateTransactionStatus generates a random transaction status
func (g *BaseGenerator) generateTransactionStatus() string {
	statuses := []string{
		"pending",
		"completed",
		"failed",
		"cancelled",
		"refunded",
	}
	return g.faker.RandomString(statuses)
}

// generateAirline generates a random airline name
func (g *BaseGenerator) generateAirline() string {
	airlines := []string{
		"Delta Airlines",
		"United Airlines",
		"American Airlines",
		"Southwest Airlines",
		"JetBlue Airways",
		"Alaska Airlines",
		"Spirit Airlines",
		"Frontier Airlines",
		"Allegiant Air",
		"Hawaiian Airlines",
	}
	return g.faker.RandomString(airlines)
}

// generateAirport generates a random airport name
func (g *BaseGenerator) generateAirport() string {
	airports := []string{
		"John F. Kennedy International Airport",
		"Los Angeles International Airport",
		"O'Hare International Airport",
		"Dallas/Fort Worth International Airport",
		"Denver International Airport",
		"San Francisco International Airport",
		"Seattle-Tacoma International Airport",
		"McCarran International Airport",
		"Orlando International Airport",
		"Charlotte Douglas International Airport",
	}
	return g.faker.RandomString(airports)
}

// generateFlightNumber generates a random flight number
func (g *BaseGenerator) generateFlightNumber() string {
	airlines := []string{"AA", "DL", "UA", "SW", "B6", "AS", "NK", "F9", "G4", "HA"}
	return g.faker.RandomString(airlines) + g.faker.RandomString([]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"})
}

// generateAircraftType generates a random aircraft type
func (g *BaseGenerator) generateAircraftType() string {
	types := []string{
		"Boeing 737",
		"Boeing 747",
		"Boeing 757",
		"Boeing 767",
		"Boeing 777",
		"Boeing 787",
		"Airbus A320",
		"Airbus A330",
		"Airbus A350",
		"Airbus A380",
	}
	return g.faker.RandomString(types)
}

// generateAircraftRegistration generates a random aircraft registration number
func (g *BaseGenerator) generateAircraftRegistration() string {
	return g.faker.RandomString([]string{"N", "C", "G", "F", "D"}) + g.faker.RandomString([]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"})
}

// generateAirportCode generates a random airport code
func (g *BaseGenerator) generateAirportCode() string {
	codes := []string{"JFK", "LAX", "ORD", "DFW", "DEN", "SFO", "SEA", "LAS", "MCO", "CLT"}
	return g.faker.RandomString(codes)
}

// generateFlightStatus generates a random flight status
func (g *BaseGenerator) generateFlightStatus() string {
	statuses := []string{
		"scheduled",
		"delayed",
		"cancelled",
		"diverted",
		"arrived",
		"departed",
	}
	return g.faker.RandomString(statuses)
}
