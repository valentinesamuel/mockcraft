package registry

import (
	"fmt"
	"sync"
)

// GeneratorFunc is a function that generates data based on given parameters
type GeneratorFunc func(params map[string]interface{}) (interface{}, error)

// IndustryRegistry manages generators for different industries
type IndustryRegistry struct {
	mu         sync.RWMutex
	generators map[string]map[string]GeneratorFunc // industry -> {generatorName -> generatorFunc}
}

// NewIndustryRegistry creates a new industry registry
func NewIndustryRegistry() *IndustryRegistry {
	return &IndustryRegistry{
		generators: make(map[string]map[string]GeneratorFunc),
	}
}

// RegisterGenerator registers a generator function for a specific industry and generator name
func (r *IndustryRegistry) RegisterGenerator(industry, generatorName string, generator GeneratorFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if industry == "" {
		return fmt.Errorf("industry name cannot be empty")
	}
	if generatorName == "" {
		return fmt.Errorf("generator name cannot be empty")
	}
	if generator == nil {
		return fmt.Errorf("generator function cannot be nil")
	}

	// Initialize industry map if it doesn't exist
	if _, exists := r.generators[industry]; !exists {
		r.generators[industry] = make(map[string]GeneratorFunc)
	}

	// Check if generator already exists
	if _, exists := r.generators[industry][generatorName]; exists {
		return fmt.Errorf("generator %s already exists for industry %s", generatorName, industry)
	}

	r.generators[industry][generatorName] = generator
	return nil
}

// GetGenerator retrieves a generator function for a specific industry and generator name
func (r *IndustryRegistry) GetGenerator(industry, generatorName string) (GeneratorFunc, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	industryGenerators, exists := r.generators[industry]
	if !exists {
		return nil, fmt.Errorf("no generators found for industry %s with generator: %s", industry, generatorName)
	}

	generator, exists := industryGenerators[generatorName]
	if !exists {
		return nil, fmt.Errorf("generator %s not found for industry %s", generatorName, industry)
	}

	return generator, nil
}

// ListIndustries returns a list of all registered industries
func (r *IndustryRegistry) ListIndustries() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	industries := make([]string, 0, len(r.generators))
	for industry := range r.generators {
		industries = append(industries, industry)
	}
	return industries
}

// ListGenerators returns a list of all generators for a specific industry
func (r *IndustryRegistry) ListGenerators(industry string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	industryGenerators, exists := r.generators[industry]
	if !exists {
		return nil, fmt.Errorf("no generators found for industry %s", industry)
	}

	generators := make([]string, 0, len(industryGenerators))
	for generator := range industryGenerators {
		generators = append(generators, generator)
	}
	return generators, nil
}

// HasGenerator checks if a generator exists for a specific industry and generator name
func (r *IndustryRegistry) HasGenerator(industry, generatorName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	industryGenerators, exists := r.generators[industry]
	if !exists {
		return false
	}

	_, exists = industryGenerators[generatorName]
	return exists
}
