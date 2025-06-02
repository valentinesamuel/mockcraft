package generators

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/valentinesamuel/mockcraft/internal/generators/interfaces"
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

// GenerateByType generates data of the specified type
func (g *BaseGenerator) GenerateByType(dataType string, params map[string]interface{}) (interface{}, error) {
	// TODO: Implement type mapping and parameter handling
	return nil, nil
}

// SetSeed sets the random seed for reproducible results
func (g *BaseGenerator) SetSeed(seed int64) {
	g.faker = gofakeit.New(seed)
}
