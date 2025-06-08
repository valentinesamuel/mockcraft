package registry

import (
	"fmt"
	"sort"
	"sync"

	"github.com/valentinesamuel/mockcraft/internal/generators/interfaces"
)

type GeneratorConstructor func() interfaces.Generator

type Registry struct {
	mu         sync.RWMutex
	generators map[string]GeneratorConstructor
}

var defaultRegistry = &Registry{
	generators: make(map[string]GeneratorConstructor),
}

// Register adds a new generator to the registry
func Register(name string, constructor GeneratorConstructor) {
	defaultRegistry.Register(name, constructor)
}

func GetAvailableGenerators() []string {
	return defaultRegistry.GetAvailableGenerators()
}

func GetDefaultGenerator() (interfaces.Generator, error) {
	return defaultRegistry.CreateGenerator("base")
}

func CreateGenerator(name string) (interfaces.Generator, error) {
	return defaultRegistry.CreateGenerator(name)
}

func (r *Registry) Register(name string, constructor GeneratorConstructor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generators[name] = constructor
}

func (r *Registry) GetAvailableGenerators() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name := range r.generators {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (r *Registry) CreateGenerator(name string) (interfaces.Generator, error) {
	r.mu.Lock()
	constructor, exists := r.generators[name]
	r.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("generator %s not found", name)
	}

	return constructor(), nil

}

func (r *Registry) IsRegistered(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.generators[name]
	return exists
}
