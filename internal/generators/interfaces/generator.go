package interfaces

// Generator defines the interface for all generators
type Generator interface {
	// GenerateByType generates data of the specified type
	GenerateByType(dataType string, params map[string]interface{}) (interface{}, error)

	// GetAvailableTypes returns all available types that can be generated
	GetAvailableTypes() []string
}
