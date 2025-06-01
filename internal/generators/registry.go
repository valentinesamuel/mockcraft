package generators

import (
	"fmt"
	"sync"

	"github.com/valentinesamuel/mockcraft/internal/generators/base"
	"github.com/valentinesamuel/mockcraft/internal/generators/health"
)

// Registry manages all available generators
type Registry struct {
	generators map[string]Generator
	mu         sync.RWMutex
}

// NewRegistry creates a new registry instance
func NewRegistry() *Registry {
	registry := &Registry{
		generators: make(map[string]Generator),
	}

	// Register built-in generators
	registry.Register("base", base.NewBaseGenerator())
	registry.Register("health", health.NewMedicalGenerator())

	return registry
}

// Register adds a new generator to the registry
func (r *Registry) Register(name string, generator Generator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generators[name] = generator
}

// Get retrieves a generator by name
func (r *Registry) Get(name string) (Generator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if generator, exists := r.generators[name]; exists {
		return generator, nil
	}
	return nil, fmt.Errorf("generator not found: %s", name)
}

// List returns all registered generator names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.generators))
	for name := range r.generators {
		names = append(names, name)
	}
	return names
}

// GetAllTypes returns all available types across all generators
func (r *Registry) GetAllTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allTypes []string
	for _, gen := range r.generators {
		allTypes = append(allTypes, gen.GetAvailableTypes()...)
	}
	return allTypes
}

var (
	registry     *Registry
	registryOnce sync.Once
)

// InitializeRegistry initializes the generator registry with all available generators
func InitializeRegistry() *Registry {
	registryOnce.Do(func() {
		registry = NewRegistry()
	})
	return registry
}

// GetRegistry returns the initialized registry
func GetRegistry() *Registry {
	if registry == nil {
		return InitializeRegistry()
	}
	return registry
}

// Generate generates data using the specified generator and type
func Generate(generatorName, dataType string, params map[string]interface{}) (interface{}, error) {
	// If no generator is specified, try all generators
	if generatorName == "" {
		for _, gen := range GetRegistry().generators {
			result, err := gen.GenerateByType(dataType, params)
			if err == nil && result != nil {
				return result, nil
			}
		}
		return nil, fmt.Errorf("type '%s' not found in any generator", dataType)
	}

	gen, err := GetRegistry().Get(generatorName)
	if err != nil {
		return nil, err
	}

	return gen.GenerateByType(dataType, params)
}

// GetAvailableTypes returns all available types for a specific generator
func GetAvailableTypes(generatorName string) ([]string, error) {
	gen, err := GetRegistry().Get(generatorName)
	if err != nil {
		return nil, err
	}

	return gen.GetAvailableTypes(), nil
}

// GetAllAvailableTypes returns all available types across all generators
func GetAllAvailableTypes() []string {
	return GetRegistry().GetAllTypes()
}

// ListGenerators returns a list of all registered generator names
func ListGenerators() []string {
	return GetRegistry().List()
}
