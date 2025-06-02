package generators

import (
	"fmt"
	"math/rand"
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
	Register("date", &DateGenerator{})
	Register("number", &NumberGenerator{})
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
	return gofakeit.DateRange(start, end), nil
}

// NumberGenerator generates numbers
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
