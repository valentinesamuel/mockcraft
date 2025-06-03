package generators

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

// Generator is the interface that all data generators must implement
type Generator interface {
	// Generate produces a single value based on the generator's type and parameters
	Generate(params map[string]interface{}) (interface{}, error)
}

var (
	registry     = make(map[string]Generator)
	registryLock sync.RWMutex
)

func init() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Register basic generators
	Register("uuid", &UUIDGenerator{})
	Register("firstname", &FirstNameGenerator{})
	Register("lastname", &LastNameGenerator{})
	Register("email", &EmailGenerator{})
	Register("phone", &PhoneGenerator{})
	Register("address", &AddressGenerator{})
	Register("company", &CompanyGenerator{})
	Register("job_title", &JobTitleGenerator{})

	// Register date/time generators
	Register("date", &DateGenerator{})
	Register("datetime", &DateTimeGenerator{})
	Register("time", &TimeGenerator{})
	Register("timestamp", &TimestampGenerator{})

	// Register number generators
	Register("number", &NumberGenerator{})
	Register("float", &FloatGenerator{})
	Register("decimal", &DecimalGenerator{})
	Register("boolean", &BooleanGenerator{})

	// Register text generators
	Register("text", &TextGenerator{})
	Register("paragraph", &ParagraphGenerator{})
	Register("sentence", &SentenceGenerator{})
	Register("word", &WordGenerator{})
	Register("char", &CharGenerator{})

	// Register internet generators
	Register("url", &URLGenerator{})
	Register("ip", &IPGenerator{})
	Register("domain", &DomainGenerator{})
	Register("username", &UsernameGenerator{})

	// Register business generators
	Register("credit_card", &CreditCardGenerator{})
	Register("currency", &CurrencyGenerator{})
	Register("price", &PriceGenerator{})
	Register("product", &ProductGenerator{})
}

// Register adds a generator to the registry
func Register(name string, generator Generator) error {
	registryLock.Lock()
	defer registryLock.Unlock()

	if _, exists := registry[name]; exists {
		return fmt.Errorf("generator '%s' already registered", name)
	}

	registry[name] = generator
	return nil
}

// Get retrieves a generator from the registry
func Get(name string) (Generator, error) {
	registryLock.RLock()
	defer registryLock.RUnlock()

	generator, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("generator '%s' not found", name)
	}

	return generator, nil
}

// List returns all registered generator names
func List() []string {
	registryLock.RLock()
	defer registryLock.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// UUIDGenerator generates UUIDs
type UUIDGenerator struct{}

func (g *UUIDGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.UUID(), nil
}

// FirstNameGenerator generates first names
type FirstNameGenerator struct{}

func (g *FirstNameGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.FirstName(), nil
}

// LastNameGenerator generates last names
type LastNameGenerator struct{}

func (g *LastNameGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.LastName(), nil
}

// EmailGenerator generates email addresses
type EmailGenerator struct{}

func (g *EmailGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.Email(), nil
}

// PhoneGenerator generates phone numbers
type PhoneGenerator struct{}

func (g *PhoneGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.Phone(), nil
}

// AddressGenerator generates addresses
type AddressGenerator struct{}

func (g *AddressGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.Address().Address, nil
}

// CompanyGenerator generates company names
type CompanyGenerator struct{}

func (g *CompanyGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.Company(), nil
}

// JobTitleGenerator generates job titles
type JobTitleGenerator struct{}

func (g *JobTitleGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.JobTitle(), nil
}

// DateGenerator generates dates
type DateGenerator struct{}

func (g *DateGenerator) Generate(params map[string]interface{}) (interface{}, error) {
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

	return gofakeit.DateRange(start, end).Format("2006-01-02"), nil
}

// DateTimeGenerator generates date and time
type DateTimeGenerator struct{}

func (g *DateTimeGenerator) Generate(params map[string]interface{}) (interface{}, error) {
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

	return gofakeit.DateRange(start, end).Format(time.RFC3339), nil
}

// TimeGenerator generates time values
type TimeGenerator struct{}

func (g *TimeGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.Date().Format("15:04:05"), nil
}

// TimestampGenerator generates Unix timestamps
type TimestampGenerator struct{}

func (g *TimestampGenerator) Generate(params map[string]interface{}) (interface{}, error) {
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

	return gofakeit.DateRange(start, end).Unix(), nil
}

// NumberGenerator generates integers
type NumberGenerator struct{}

func (g *NumberGenerator) Generate(params map[string]interface{}) (interface{}, error) {
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

	return rand.Intn(max-min+1) + min, nil
}

// FloatGenerator generates floating-point numbers
type FloatGenerator struct{}

func (g *FloatGenerator) Generate(params map[string]interface{}) (interface{}, error) {
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
	return fmt.Sprintf("%.*f", precision, value), nil
}

// DecimalGenerator generates decimal numbers
type DecimalGenerator struct{}

func (g *DecimalGenerator) Generate(params map[string]interface{}) (interface{}, error) {
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
	return fmt.Sprintf("%.*f", precision, value), nil
}

// BooleanGenerator generates boolean values
type BooleanGenerator struct{}

func (g *BooleanGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return rand.Float32() < 0.5, nil
}

// TextGenerator generates random text
type TextGenerator struct{}

func (g *TextGenerator) Generate(params map[string]interface{}) (interface{}, error) {
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
		chars[i] = gofakeit.RandomString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})
	}
	return strings.Join(chars, ""), nil
}

// ParagraphGenerator generates paragraphs
type ParagraphGenerator struct{}

func (g *ParagraphGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	count := 1
	if countVal, ok := params["count"].(int); ok {
		count = countVal
	}
	return gofakeit.Paragraph(count, count, 10, "\n"), nil
}

// SentenceGenerator generates sentences
type SentenceGenerator struct{}

func (g *SentenceGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.Sentence(10), nil
}

// WordGenerator generates words
type WordGenerator struct{}

func (g *WordGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.Word(), nil
}

// CharGenerator generates single characters
type CharGenerator struct{}

func (g *CharGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return string(gofakeit.RandomString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})[0]), nil
}

// URLGenerator generates URLs
type URLGenerator struct{}

func (g *URLGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.URL(), nil
}

// IPGenerator generates IP addresses
type IPGenerator struct{}

func (g *IPGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.IPv4Address(), nil
}

// DomainGenerator generates domain names
type DomainGenerator struct{}

func (g *DomainGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.DomainName(), nil
}

// UsernameGenerator generates usernames
type UsernameGenerator struct{}

func (g *UsernameGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.Username(), nil
}

// CreditCardGenerator generates credit card numbers
type CreditCardGenerator struct{}

func (g *CreditCardGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.CreditCardNumber(nil), nil
}

// CurrencyGenerator generates currency codes
type CurrencyGenerator struct{}

func (g *CurrencyGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.CurrencyShort(), nil
}

// PriceGenerator generates prices
type PriceGenerator struct{}

func (g *PriceGenerator) Generate(params map[string]interface{}) (interface{}, error) {
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
	return fmt.Sprintf("%.*f", precision, value), nil
}

// ProductGenerator generates product names
type ProductGenerator struct{}

func (g *ProductGenerator) Generate(params map[string]interface{}) (interface{}, error) {
	return gofakeit.ProductName(), nil
}
