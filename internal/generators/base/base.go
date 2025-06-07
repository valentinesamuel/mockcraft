package base

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/valentinesamuel/mockcraft/internal/generators/health"
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
		"bank_name",
		"account_number",
		"routing_number",
		"currency",
		"currency_code",
		"stock_symbol",
		"stock_price",
		"transaction_type",
		"transaction_status",
		"airline",
		"airport",
		"flight_number",
		"aircraft_type",
		"aircraft_registration",
		"airport_code",
		"flight_status",
		"blood_type",
		"medical_condition",
		"medication",
		"symptom",
		"diagnosis",
		"allergy",
		"lab_result",
		"vital_sign",
		"medical_record",
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

	// Handle custom types
	switch dataType {
	// Basic generators
	case "uuid":
		version := 4 // default version
		if v, ok := params["version"].(int); ok {
			version = v
		}
		result = g.generateUUID(version)

	case "first_name":
		result = g.faker.FirstName()

	case "last_name":
		result = g.faker.LastName()

	case "email":
		result = g.faker.Email()

	case "phone":
		format := "international" // default format
		if f, ok := params["format"].(string); ok {
			format = f
		}
		result = g.generatePhone(format)

	case "address":
		result = g.faker.Address().Address

	case "company":
		result = g.faker.Company()

	case "job_title":
		result = g.faker.JobTitle()

	// Date/time generators
	case "date":
		start := time.Now().AddDate(-1, 0, 0) // 1 year ago
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

		result = g.faker.DateRange(start, end).Format("2006-01-02")

	case "datetime":
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

		result = g.faker.DateRange(start, end).Format(time.RFC3339)

	case "time":
		result = g.faker.Date().Format("15:04:05")

	case "timestamp":
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

		result = g.faker.DateRange(start, end).Unix()

	// Number generators
	case "number":
		min := 0
		max := 100

		if minVal, ok := params["min"].(int); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(int); ok {
			max = maxVal
		}

		if min > max {
			return nil, fmt.Errorf("min value (%d) cannot be greater than max value (%d)", min, max)
		}

		result = rand.Intn(max-min+1) + min

	case "float":
		min := 0.0
		max := 100.0
		precision := 2

		if minVal, ok := params["min"].(float64); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(float64); ok {
			max = maxVal
		}
		if prec, ok := params["precision"].(int); ok {
			precision = prec
		}

		if min > max {
			return nil, fmt.Errorf("min value (%f) cannot be greater than max value (%f)", min, max)
		}

		value := min + rand.Float64()*(max-min)
		result = fmt.Sprintf("%.*f", precision, value)

	case "decimal":
		min := 0.0
		max := 100.0
		precision := 2

		if minVal, ok := params["min"].(float64); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(float64); ok {
			max = maxVal
		}
		if prec, ok := params["precision"].(int); ok {
			precision = prec
		}

		if min > max {
			return nil, fmt.Errorf("min value (%f) cannot be greater than max value (%f)", min, max)
		}

		value := min + rand.Float64()*(max-min)
		result = fmt.Sprintf("%.*f", precision, value)

	case "boolean":
		result = rand.Float32() < 0.5

	// Text generators
	case "text":
		min := 10
		max := 100

		if minVal, ok := params["min"].(int); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(int); ok {
			max = maxVal
		}

		if min > max {
			return nil, fmt.Errorf("min length (%d) cannot be greater than max length (%d)", min, max)
		}

		length := min + rand.Intn(max-min+1)
		chars := make([]string, length)
		for i := 0; i < length; i++ {
			chars[i] = g.faker.RandomString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})
		}
		result = strings.Join(chars, "")

	case "paragraph":
		count := 1
		if countVal, ok := params["count"].(int); ok {
			count = countVal
		}
		result = g.faker.Paragraph(count, count, 10, "\n")

	case "sentence":
		result = g.faker.Sentence(10)

	case "word":
		result = g.faker.Word()

	case "char":
		result = string(g.faker.RandomString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})[0])

	// Internet generators
	case "url":
		result = g.faker.URL()

	case "ip":
		result = g.faker.IPv4Address()

	case "domain":
		result = g.faker.DomainName()

	case "username":
		result = g.faker.Username()

	// Business generators
	case "credit_card":
		result = g.faker.CreditCardNumber(nil)

	case "currency":
		result = g.faker.CurrencyShort()

	case "price":
		min := 0.0
		max := 1000.0
		precision := 2

		if minVal, ok := params["min"].(float64); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(float64); ok {
			max = maxVal
		}
		if prec, ok := params["precision"].(int); ok {
			precision = prec
		}

		if min > max {
			return nil, fmt.Errorf("min value (%f) cannot be greater than max value (%f)", min, max)
		}

		value := min + rand.Float64()*(max-min)
		result = fmt.Sprintf("%.*f", precision, value)

	case "product":
		result = g.faker.ProductName()

	// Finance generators
	case "bank_name":
		result = g.generateBankName()
	case "account_number":
		result = g.generateAccountNumber()
	case "routing_number":
		result = g.generateRoutingNumber()
	case "currency_code":
		result = g.generateCurrencyCode()
	case "stock_symbol":
		result = g.generateStockSymbol()
	case "stock_price":
		result = g.generateStockPrice()
	case "transaction_type":
		result = g.generateTransactionType()
	case "transaction_status":
		result = g.generateTransactionStatus()

	// Aviation generators
	case "airline":
		result = g.generateAirline()
	case "airport":
		result = g.generateAirport()
	case "flight_number":
		result = g.generateFlightNumber()
	case "aircraft_type":
		result = g.generateAircraftType()
	case "aircraft_registration":
		result = g.generateAircraftRegistration()
	case "airport_code":
		result = g.generateAirportCode()
	case "flight_status":
		result = g.generateFlightStatus()

	// Health generators
	case "blood_type", "medical_condition", "medication", "symptom", "diagnosis", "allergy", "lab_result", "vital_sign", "medical_record":
		return g.generateHealthType(dataType)

	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}

	// Apply text transformations if applicable
	return g.applyTextTransformations(result, params), nil
}

// validateParameters validates the parameters for a given generator type
func (g *BaseGenerator) validateParameters(dataType string, params map[string]interface{}) error {
	switch dataType {
	case "number", "float", "decimal", "price":
		if min, ok := params["min"].(float64); ok {
			if max, ok := params["max"].(float64); ok {
				if min > max {
					return fmt.Errorf("min value (%f) cannot be greater than max value (%f)", min, max)
				}
			}
		}

	case "text":
		if min, ok := params["min"].(int); ok {
			if max, ok := params["max"].(int); ok {
				if min > max {
					return fmt.Errorf("min length (%d) cannot be greater than max length (%d)", min, max)
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
					return fmt.Errorf("start time (%s) cannot be after end time (%s)", start, end)
				}
			}
		}
	}

	return nil
}

// generateUUID generates a UUID of the specified version
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

// generatePhone generates a phone number in the specified format
func (g *BaseGenerator) generatePhone(format string) string {
	switch format {
	case "international":
		return g.faker.Phone()
	case "national":
		return g.faker.Phone()
	case "local":
		return g.faker.Phone()
	default:
		return g.faker.Phone()
	}
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

// Finance generator functions
func (g *BaseGenerator) generateBankName() string {
	banks := []string{
		"JPMorgan Chase", "Bank of America", "Wells Fargo", "Citibank",
		"Goldman Sachs", "Morgan Stanley", "HSBC", "Barclays",
		"Deutsche Bank", "UBS", "Credit Suisse", "BNP Paribas",
	}
	return banks[rand.Intn(len(banks))]
}

func (g *BaseGenerator) generateAccountNumber() string {
	number := ""
	for i := 0; i < 10; i++ {
		number += fmt.Sprintf("%d", rand.Intn(10))
	}
	return number
}

func (g *BaseGenerator) generateRoutingNumber() string {
	number := ""
	for i := 0; i < 9; i++ {
		number += fmt.Sprintf("%d", rand.Intn(10))
	}
	return number
}

func (g *BaseGenerator) generateCreditCard() string {
	return g.faker.CreditCardNumber(nil)
}

func (g *BaseGenerator) generateCurrency() string {
	currencies := []string{
		"US Dollar", "Euro", "British Pound", "Japanese Yen",
		"Swiss Franc", "Canadian Dollar", "Australian Dollar",
		"Chinese Yuan", "Indian Rupee", "Brazilian Real",
	}
	return currencies[rand.Intn(len(currencies))]
}

func (g *BaseGenerator) generateCurrencyCode() string {
	codes := []string{
		"USD", "EUR", "GBP", "JPY", "CHF", "CAD", "AUD", "CNY", "INR", "BRL",
	}
	return codes[rand.Intn(len(codes))]
}

func (g *BaseGenerator) generateStockSymbol() string {
	symbols := []string{
		"AAPL", "MSFT", "GOOGL", "AMZN", "META", "TSLA", "NVDA", "JPM",
		"V", "WMT", "PG", "JNJ", "MA", "UNH", "HD", "BAC", "XOM", "DIS",
	}
	return symbols[rand.Intn(len(symbols))]
}

func (g *BaseGenerator) generateStockPrice() float64 {
	return 10 + rand.Float64()*990
}

func (g *BaseGenerator) generateTransactionType() string {
	types := []string{
		"Deposit", "Withdrawal", "Transfer", "Payment",
		"Purchase", "Refund", "Fee", "Interest",
	}
	return types[rand.Intn(len(types))]
}

func (g *BaseGenerator) generateTransactionStatus() string {
	statuses := []string{
		"Completed", "Pending", "Failed", "Reversed",
		"Processing", "Cancelled", "Refunded",
	}
	return statuses[rand.Intn(len(statuses))]
}

// Aviation generator functions
func (g *BaseGenerator) generateAirline() string {
	airlines := []string{
		"Delta Air Lines", "American Airlines", "United Airlines", "Southwest Airlines",
		"British Airways", "Lufthansa", "Air France", "Emirates",
		"Qatar Airways", "Singapore Airlines", "Cathay Pacific", "ANA",
	}
	return airlines[rand.Intn(len(airlines))]
}

func (g *BaseGenerator) generateAirport() string {
	airports := []string{
		"John F. Kennedy International", "Los Angeles International", "Heathrow Airport",
		"Charles de Gaulle Airport", "Dubai International", "Singapore Changi",
		"Hong Kong International", "Tokyo Haneda", "Sydney Airport", "Amsterdam Schiphol",
	}
	return airports[rand.Intn(len(airports))]
}

func (g *BaseGenerator) generateFlightNumber() string {
	airline := g.generateAirline()
	number := rand.Intn(9999)
	return fmt.Sprintf("%s%d", airline[:2], number)
}

func (g *BaseGenerator) generateAircraftType() string {
	types := []string{
		"Boeing 737", "Boeing 747", "Boeing 777", "Boeing 787",
		"Airbus A320", "Airbus A330", "Airbus A350", "Airbus A380",
		"Embraer E190", "Bombardier CRJ900", "ATR 72", "Dash 8",
	}
	return types[rand.Intn(len(types))]
}

func (g *BaseGenerator) generateAircraftRegistration() string {
	prefixes := []string{"N", "G", "F", "D", "B", "JA", "VH", "C"}
	prefix := prefixes[rand.Intn(len(prefixes))]
	number := rand.Intn(999999)
	return fmt.Sprintf("%s-%d", prefix, number)
}

func (g *BaseGenerator) generateAirportCode() string {
	codes := []string{
		"JFK", "LAX", "LHR", "CDG", "DXB", "SIN", "HKG", "HND", "SYD", "AMS",
	}
	return codes[rand.Intn(len(codes))]
}

func (g *BaseGenerator) generateFlightStatus() string {
	statuses := []string{
		"Scheduled", "Boarding", "In Flight", "Landed",
		"Delayed", "Cancelled", "Diverted", "On Time",
	}
	return statuses[rand.Intn(len(statuses))]
}
