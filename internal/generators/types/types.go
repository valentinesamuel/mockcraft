package types

// Category represents a group of related data types
type Category struct {
	Name        string
	Description string
	Types       []Type
}

// Type represents a data type that can be generated
type Type struct {
	Name        string
	Description string
	Example     string
	Parameters  []Parameter
	Category    string
}

// Parameter represents a parameter that can be passed to a type
type Parameter struct {
	Name        string
	Description string
	Type        string
	Required    bool
	Default     interface{}
}

// HealthTypes maps health-related type names to their descriptions
var HealthTypes = map[string]string{
	"systolic":            "Systolic blood pressure (90-140 mmHg)",
	"diastolic":           "Diastolic blood pressure (60-90 mmHg)",
	"blood_pressure_unit": "Blood pressure unit (mmHg)",
	"heart_rate":          "Heart rate (60-100 bpm)",
	"heart_rate_unit":     "Heart rate unit (bpm)",
	"temperature":         "Body temperature (97.0-99.0 °F)",
	"temperature_unit":    "Temperature unit (°F)",
	"respiratory_rate":    "Respiratory rate (12-20 breaths/min)",
	"respiratory_unit":    "Respiratory rate unit (breaths/min)",
}

// GetAllCategories returns all available categories
func GetAllCategories() []Category {
	return []Category{
		{
			Name:        "personal",
			Description: "Personal information like names, addresses, etc.",
			Types: []Type{
				{
					Name:        "first_name",
					Description: "Generate a random first name",
					Example:     "John",
					Category:    "personal",
				},
				{
					Name:        "last_name",
					Description: "Generate a random last name",
					Example:     "Doe",
					Category:    "personal",
				},
				{
					Name:        "email",
					Description: "Generate a random email address",
					Example:     "john.doe@example.com",
					Category:    "personal",
				},
				{
					Name:        "phone",
					Description: "Generate a random phone number",
					Example:     "+1 (555) 123-4567",
					Category:    "personal",
					Parameters: []Parameter{
						{
							Name:        "format",
							Description: "Phone number format (international, national, local)",
							Type:        "string",
							Required:    false,
							Default:     "international",
						},
					},
				},
				{
					Name:        "address",
					Description: "Generate a random address",
					Example:     "123 Main St, Anytown, CA 12345",
					Category:    "personal",
					Parameters: []Parameter{
						{
							Name:        "country",
							Description: "Country code (e.g., US, UK, CA)",
							Type:        "string",
							Required:    false,
							Default:     "US",
						},
					},
				},
				{
					Name:        "ssn",
					Description: "Generate a random Social Security Number",
					Example:     "123-45-6789",
					Category:    "personal",
				},
			},
		},
		{
			Name:        "technical",
			Description: "Technical data like passwords, IP addresses, etc.",
			Types: []Type{
				{
					Name:        "password",
					Description: "Generate a secure random password",
					Example:     "xK9#mP2$vL5nQ8@j",
					Category:    "technical",
					Parameters: []Parameter{
						{
							Name:        "length",
							Description: "Length of the password",
							Type:        "int",
							Required:    false,
							Default:     12,
						},
					},
				},
				{
					Name:        "ip_address",
					Description: "Generate a random IP address",
					Example:     "192.168.1.1",
					Category:    "technical",
					Parameters: []Parameter{
						{
							Name:        "version",
							Description: "IP version (4 or 6)",
							Type:        "int",
							Required:    false,
							Default:     4,
						},
					},
				},
				{
					Name:        "uuid",
					Description: "Generate a random UUID",
					Example:     "123e4567-e89b-12d3-a456-426614174000",
					Category:    "technical",
					Parameters: []Parameter{
						{
							Name:        "version",
							Description: "UUID version (1, 3, 4, 5)",
							Type:        "int",
							Required:    false,
							Default:     4,
						},
					},
				},
				{
					Name:        "mac_address",
					Description: "Generate a random MAC address",
					Example:     "00:1A:2B:3C:4D:5E",
					Category:    "technical",
				},
				{
					Name:        "domain",
					Description: "Generate a random domain name",
					Example:     "example.com",
					Category:    "technical",
					Parameters: []Parameter{
						{
							Name:        "tld",
							Description: "Top-level domain (com, org, net, etc.)",
							Type:        "string",
							Required:    false,
							Default:     "com",
						},
					},
				},
			},
		},
		{
			Name:        "text",
			Description: "Text generation like sentences, paragraphs, etc.",
			Types: []Type{
				{
					Name:        "sentence",
					Description: "Generate a random sentence",
					Example:     "The quick brown fox jumps over the lazy dog.",
					Category:    "text",
					Parameters: []Parameter{
						{
							Name:        "word_count",
							Description: "Number of words in the sentence",
							Type:        "int",
							Required:    false,
							Default:     6,
						},
					},
				},
				{
					Name:        "paragraph",
					Description: "Generate a random paragraph",
					Example:     "Lorem ipsum dolor sit amet...",
					Category:    "text",
					Parameters: []Parameter{
						{
							Name:        "sentence_count",
							Description: "Number of sentences in the paragraph",
							Type:        "int",
							Required:    false,
							Default:     3,
						},
					},
				},
				{
					Name:        "word",
					Description: "Generate a random word",
					Example:     "example",
					Category:    "text",
					Parameters: []Parameter{
						{
							Name:        "min_length",
							Description: "Minimum word length",
							Type:        "int",
							Required:    false,
							Default:     4,
						},
						{
							Name:        "max_length",
							Description: "Maximum word length",
							Type:        "int",
							Required:    false,
							Default:     8,
						},
					},
				},
			},
		},
		{
			Name:        "utility",
			Description: "Utility functions for data manipulation",
			Types: []Type{
				{
					Name:        "shuffle_strings",
					Description: "Shuffle a comma-separated list of strings",
					Example:     "banana,apple,cherry",
					Category:    "utility",
					Parameters: []Parameter{
						{
							Name:        "strings",
							Description: "Comma-separated list of strings to shuffle",
							Type:        "string",
							Required:    true,
						},
					},
				},
				{
					Name:        "random_int",
					Description: "Generate a random integer within a range",
					Example:     "42",
					Category:    "utility",
					Parameters: []Parameter{
						{
							Name:        "min",
							Description: "Minimum value (inclusive)",
							Type:        "int",
							Required:    true,
						},
						{
							Name:        "max",
							Description: "Maximum value (inclusive)",
							Type:        "int",
							Required:    true,
						},
					},
				},
				{
					Name:        "random_float",
					Description: "Generate a random float within a range",
					Example:     "3.14159",
					Category:    "utility",
					Parameters: []Parameter{
						{
							Name:        "min",
							Description: "Minimum value (inclusive)",
							Type:        "float",
							Required:    true,
						},
						{
							Name:        "max",
							Description: "Maximum value (inclusive)",
							Type:        "float",
							Required:    true,
						},
						{
							Name:        "precision",
							Description: "Number of decimal places",
							Type:        "int",
							Required:    false,
							Default:     2,
						},
					},
				},
			},
		},
	}
}

// GetTypeByName returns a type by its name
func GetTypeByName(name string) *Type {
	for _, cat := range GetAllCategories() {
		for _, t := range cat.Types {
			if t.Name == name {
				return &t
			}
		}
	}
	return nil
}

// GetTypesByCategory returns all types in a category
func GetTypesByCategory(categoryName string) []Type {
	for _, cat := range GetAllCategories() {
		if cat.Name == categoryName {
			return cat.Types
		}
	}
	return nil
}

// GetAllTypes returns all available types
func GetAllTypes() []Type {
	var types []Type
	for _, cat := range GetAllCategories() {
		types = append(types, cat.Types...)
	}
	return types
}
