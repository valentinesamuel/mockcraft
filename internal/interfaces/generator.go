package interfaces

// Generator defines the interface for data generators
type Generator interface {
	GenerateByType(dataType string, params map[string]interface{}) (interface{}, error)
	GetAvailableTypes() []string
	SetSeed(seed int64)
}
