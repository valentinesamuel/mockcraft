package interfaces

// Generator defines the interface for all data generators
type Generator interface {
	// GenerateByType generates data of the specified type
	GenerateByType(dataType string, params map[string]interface{}) (interface{}, error)

	// SetSeed sets the random seed for reproducible results
	SetSeed(seed int64)

	// GetAvailableTypes returns a list of all available data types
	GetAvailableTypes() []string
}
