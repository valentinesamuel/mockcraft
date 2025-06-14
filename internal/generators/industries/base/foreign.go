package base

import (
	"fmt"
	"sync"

	"github.com/valentinesamuel/mockcraft/internal/generators/interfaces"
	"github.com/valentinesamuel/mockcraft/internal/registry"
)

var (
	foreignValues = make(map[string][]interface{})
	foreignMutex  sync.RWMutex
)

// RegisterForeignValues registers values for a foreign key relationship
func RegisterForeignValues(table, column string, values []interface{}) {
	key := fmt.Sprintf("%s.%s", table, column)
	foreignMutex.Lock()
	defer foreignMutex.Unlock()
	foreignValues[key] = values
}

// GetForeignValues gets values for a foreign key relationship
func GetForeignValues(table, column string) []interface{} {
	key := fmt.Sprintf("%s.%s", table, column)
	foreignMutex.RLock()
	defer foreignMutex.RUnlock()
	return foreignValues[key]
}

// ForeignGenerator implements the Generator interface for foreign keys
type ForeignGenerator struct{}

// NewForeignGenerator creates a new ForeignGenerator instance
func NewForeignGenerator() interfaces.Generator {
	return &ForeignGenerator{}
}

// Auto-register when package is imported
func init() {
	registry.Register("foreign", NewForeignGenerator)
}

// GetAvailableTypes returns the list of available generator types
func (g *ForeignGenerator) GetAvailableTypes() []string {
	return []string{"foreign"}
}

// GenerateByType generates data of the specified type
func (g *ForeignGenerator) GenerateByType(dataType string, params map[string]interface{}) (interface{}, error) {
	if dataType != "foreign" {
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}

	table, ok := params["table"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required parameter 'table'")
	}

	column, ok := params["column"].(string)
	if !ok {
		return nil, fmt.Errorf("missing required parameter 'column'")
	}

	values := GetForeignValues(table, column)
	if len(values) == 0 {
		return nil, fmt.Errorf("no values found for foreign key %s.%s", table, column)
	}

	// Use the first value for now - in a real implementation, you'd want to randomly select one
	return values[0], nil
}

// SetSeed sets the random seed for reproducible results
func (g *ForeignGenerator) SetSeed(seed int64) {
	// No-op for foreign generator
}
