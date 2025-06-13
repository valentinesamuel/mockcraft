package types

// TypeDefinition defines a generator type and its parameters
type TypeDefinition struct {
	Name        string
	Description string
	Example     string
	Parameters  []Parameter
	Category    string
	Industry    string
}

// Parameter defines a parameter for a generator type
type Parameter struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     interface{}
}

// HealthTypes defines all available health-related generator types
var HealthTypes = map[string]TypeDefinition{
	"blood_type": {
		Name:        "blood_type",
		Description: "Generate a random blood type (A+, A-, B+, B-, AB+, AB-, O+, O-)",
		Example:     "A+",
		Industry:    "health",
	},
	"medical_condition": {
		Name:        "medical_condition",
		Description: "Generate a random medical condition",
		Example:     "Hypertension",
		Industry:    "health",
	},
	"medication": {
		Name:        "medication",
		Description: "Generate a random medication name",
		Example:     "Lisinopril",
		Industry:    "health",
	},
}

// GetAllTypes returns all available types
func GetAllTypes() []TypeDefinition {
	var allTypes []TypeDefinition

	// Add health types
	for _, t := range HealthTypes {
		allTypes = append(allTypes, t)
	}

	// Add more types here as they are implemented
	return allTypes
}

// GetTypeByName returns a type definition by name
func GetTypeByName(name string) *TypeDefinition {
	// Check health types
	if t, ok := HealthTypes[name]; ok {
		return &t
	}

	// Add more type categories here as they are implemented
	return nil
}

// GetTypesByCategory returns all types in a category
func GetTypesByCategory(category string) []TypeDefinition {
	var types []TypeDefinition

	// Check health types
	if category == "health" {
		for _, t := range HealthTypes {
			types = append(types, t)
		}
	}

	// Add more categories here as they are implemented
	return types
}

// GetAllCategories returns all available categories
func GetAllCategories() []string {
	categories := make(map[string]bool)

	// Add health category
	categories["health"] = true

	// Add more categories here as they are implemented
	var result []string
	for cat := range categories {
		result = append(result, cat)
	}
	return result
}
