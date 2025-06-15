package generators

import "fmt"

// registerAllParameters registers parameter definitions for all generators
func (pr *ParameterRegistry) registerAllParameters() {
	pr.registerBaseParameters()
	pr.registerHealthParameters()
	pr.registerAviationParameters()
}

// registerBaseParameters registers parameters for base industry generators
func (pr *ParameterRegistry) registerBaseParameters() {
	// UUID generator
	pr.RegisterGenerator("base", "uuid", GeneratorInfo{
		Description: "Generates a random UUID (Universally Unique Identifier)",
		Example:     "550e8400-e29b-41d4-a716-446655440000",
		Parameters: []ParameterDef{
			{
				Name:        "version",
				Type:        ParamTypeSelect,
				Description: "UUID version to generate",
				Default:     "4",
				Options:     []string{"1", "4"},
				Example:     "4",
			},
		},
	})

	// First name generator
	pr.RegisterGenerator("base", "firstname", GeneratorInfo{
		Description: "Generates a random first name",
		Example:     "John",
		Parameters: []ParameterDef{
			{
				Name:        "uppercase",
				Type:        ParamTypeBool,
				Description: "Convert to uppercase",
				Default:     false,
				Example:     true,
			},
			{
				Name:        "lowercase",
				Type:        ParamTypeBool,
				Description: "Convert to lowercase",
				Default:     false,
				Example:     true,
			},
			{
				Name:        "capitalize",
				Type:        ParamTypeBool,
				Description: "Capitalize first letter",
				Default:     true,
				Example:     true,
			},
		},
	})

	// Last name generator
	pr.RegisterGenerator("base", "lastname", GeneratorInfo{
		Description: "Generates a random last name",
		Example:     "Smith",
		Parameters: []ParameterDef{
			{
				Name:        "uppercase",
				Type:        ParamTypeBool,
				Description: "Convert to uppercase",
				Default:     false,
				Example:     true,
			},
			{
				Name:        "lowercase",
				Type:        ParamTypeBool,
				Description: "Convert to lowercase",
				Default:     false,
				Example:     true,
			},
			{
				Name:        "capitalize",
				Type:        ParamTypeBool,
				Description: "Capitalize first letter",
				Default:     true,
				Example:     true,
			},
		},
	})

	// Email generator
	pr.RegisterGenerator("base", "email", GeneratorInfo{
		Description: "Generates a random email address",
		Example:     "john.doe@example.com",
		Parameters: []ParameterDef{
			{
				Name:        "domain",
				Type:        ParamTypeString,
				Description: "Specify domain for email",
				Example:     "company.com",
			},
			{
				Name:        "uppercase",
				Type:        ParamTypeBool,
				Description: "Convert to uppercase",
				Default:     false,
				Example:     false,
			},
			{
				Name:        "lowercase",
				Type:        ParamTypeBool,
				Description: "Convert to lowercase",
				Default:     true,
				Example:     true,
			},
		},
	})

	// Phone generator
	pr.RegisterGenerator("base", "phone", GeneratorInfo{
		Description: "Generates a random phone number",
		Example:     "(555) 123-4567",
		Parameters: []ParameterDef{
			{
				Name:        "format",
				Type:        ParamTypeSelect,
				Description: "Phone number format",
				Default:     "national",
				Options:     []string{"international", "national", "local"},
				Example:     "international",
			},
			{
				Name:        "country",
				Type:        ParamTypeString,
				Description: "Country code for phone number",
				Default:     "US",
				Example:     "GB",
			},
		},
	})

	// Password generator
	pr.RegisterGenerator("base", "password", GeneratorInfo{
		Description: "Generates a random secure password",
		Example:     "aB3$kL9mP2qR",
		Parameters: []ParameterDef{
			{
				Name:        "length",
				Type:        ParamTypeInt,
				Description: "Length of the password",
				Default:     12,
				Min:         4,
				Max:         128,
				Example:     16,
			},
			{
				Name:        "include_symbols",
				Type:        ParamTypeBool,
				Description: "Include special symbols",
				Default:     true,
				Example:     false,
			},
			{
				Name:        "include_numbers",
				Type:        ParamTypeBool,
				Description: "Include numbers",
				Default:     true,
				Example:     true,
			},
		},
	})

	// Number generator
	pr.RegisterGenerator("base", "number", GeneratorInfo{
		Description: "Generates a random integer number",
		Example:     "42",
		Parameters: []ParameterDef{
			{
				Name:        "min",
				Type:        ParamTypeInt,
				Description: "Minimum value",
				Default:     0,
				Example:     1,
			},
			{
				Name:        "max",
				Type:        ParamTypeInt,
				Description: "Maximum value",
				Default:     100,
				Example:     1000,
			},
			{
				Name:        "positive",
				Type:        ParamTypeBool,
				Description: "Generate only positive numbers",
				Default:     false,
				Example:     true,
			},
			{
				Name:        "unit",
				Type:        ParamTypeString,
				Description: "Unit to append to the number",
				Example:     "kg",
			},
		},
	})

	// Float generator
	pr.RegisterGenerator("base", "float", GeneratorInfo{
		Description: "Generates a random floating-point number",
		Example:     "42.56",
		Parameters: []ParameterDef{
			{
				Name:        "min",
				Type:        ParamTypeFloat,
				Description: "Minimum value",
				Default:     0.0,
				Example:     1.5,
			},
			{
				Name:        "max",
				Type:        ParamTypeFloat,
				Description: "Maximum value",
				Default:     100.0,
				Example:     999.9,
			},
			{
				Name:        "precision",
				Type:        ParamTypeInt,
				Description: "Number of decimal places",
				Default:     2,
				Min:         0,
				Max:         10,
				Example:     3,
			},
			{
				Name:        "unit",
				Type:        ParamTypeString,
				Description: "Unit to append to the number",
				Example:     "m",
			},
		},
	})

	// Text generator
	pr.RegisterGenerator("base", "text", GeneratorInfo{
		Description: "Generates random text with specified length",
		Example:     "lorem ipsum dolor",
		Parameters: []ParameterDef{
			{
				Name:        "min",
				Type:        ParamTypeInt,
				Description: "Minimum length",
				Default:     10,
				Min:         1,
				Example:     5,
			},
			{
				Name:        "max",
				Type:        ParamTypeInt,
				Description: "Maximum length",
				Default:     100,
				Min:         1,
				Example:     50,
			},
			{
				Name:        "uppercase",
				Type:        ParamTypeBool,
				Description: "Convert to uppercase",
				Default:     false,
				Example:     true,
			},
			{
				Name:        "lowercase",
				Type:        ParamTypeBool,
				Description: "Convert to lowercase",
				Default:     false,
				Example:     true,
			},
			{
				Name:        "prefix",
				Type:        ParamTypeString,
				Description: "Text to add at the beginning",
				Example:     "Mr. ",
			},
			{
				Name:        "suffix",
				Type:        ParamTypeString,
				Description: "Text to add at the end",
				Example:     "!",
			},
		},
	})

	// Sentence generator
	pr.RegisterGenerator("base", "sentence", GeneratorInfo{
		Description: "Generates a random sentence",
		Example:     "The quick brown fox jumps over the lazy dog.",
		Parameters: []ParameterDef{
			{
				Name:        "word_count",
				Type:        ParamTypeInt,
				Description: "Number of words in the sentence",
				Default:     10,
				Min:         3,
				Max:         50,
				Example:     15,
			},
			{
				Name:        "uppercase",
				Type:        ParamTypeBool,
				Description: "Convert to uppercase",
				Default:     false,
				Example:     false,
			},
			{
				Name:        "lowercase",
				Type:        ParamTypeBool,
				Description: "Convert to lowercase",
				Default:     false,
				Example:     false,
			},
		},
	})

	// Paragraph generator
	pr.RegisterGenerator("base", "paragraph", GeneratorInfo{
		Description: "Generates a random paragraph",
		Example:     "Lorem ipsum dolor sit amet. Consectetur adipiscing elit sed do.",
		Parameters: []ParameterDef{
			{
				Name:        "sentence_count",
				Type:        ParamTypeInt,
				Description: "Number of sentences in the paragraph",
				Default:     3,
				Min:         1,
				Max:         20,
				Example:     5,
			},
			{
				Name:        "word_count",
				Type:        ParamTypeInt,
				Description: "Average words per sentence",
				Default:     10,
				Min:         3,
				Max:         30,
				Example:     12,
			},
		},
	})

	// Date generator
	pr.RegisterGenerator("base", "date", GeneratorInfo{
		Description: "Generates a random date",
		Example:     "2024-03-15",
		Parameters: []ParameterDef{
			{
				Name:        "format",
				Type:        ParamTypeString,
				Description: "Date format (Go time format)",
				Default:     "2006-01-02",
				Example:     "01/02/2006",
			},
			{
				Name:        "start_date",
				Type:        ParamTypeString,
				Description: "Start date for range (YYYY-MM-DD)",
				Example:     "2024-01-01",
			},
			{
				Name:        "end_date",
				Type:        ParamTypeString,
				Description: "End date for range (YYYY-MM-DD)",
				Example:     "2024-12-31",
			},
		},
	})

	// DateTime generator
	pr.RegisterGenerator("base", "datetime", GeneratorInfo{
		Description: "Generates a random date and time",
		Example:     "2024-03-15T14:30:00Z",
		Parameters: []ParameterDef{
			{
				Name:        "format",
				Type:        ParamTypeString,
				Description: "DateTime format (Go time format)",
				Default:     "2006-01-02T15:04:05Z07:00",
				Example:     "2006-01-02 15:04:05",
			},
			{
				Name:        "timezone",
				Type:        ParamTypeString,
				Description: "Timezone for the datetime",
				Default:     "UTC",
				Example:     "America/New_York",
			},
		},
	})

	// Boolean generator
	pr.RegisterGenerator("base", "boolean", GeneratorInfo{
		Description: "Generates a random boolean value",
		Example:     "true",
		Parameters:  []ParameterDef{}, // No parameters for boolean
	})

	// Enum generator
	pr.RegisterGenerator("base", "enum", GeneratorInfo{
		Description: "Generates a random value from a user-provided list",
		Example:     "apple",
		Parameters: []ParameterDef{
			{
				Name:        "values",
				Type:        ParamTypeStringSlice,
				Description: "List of values to choose from",
				Required:    true,
				Example:     []string{"apple", "banana", "orange"},
			},
		},
	})

	// Word generator
	pr.RegisterGenerator("base", "word", GeneratorInfo{
		Description: "Generates a random word",
		Example:     "elephant",
		Parameters: []ParameterDef{
			{
				Name:        "min_length",
				Type:        ParamTypeInt,
				Description: "Minimum word length",
				Default:     3,
				Min:         1,
				Example:     5,
			},
			{
				Name:        "max_length",
				Type:        ParamTypeInt,
				Description: "Maximum word length",
				Default:     12,
				Min:         1,
				Example:     10,
			},
			{
				Name:        "uppercase",
				Type:        ParamTypeBool,
				Description: "Convert to uppercase",
				Default:     false,
				Example:     true,
			},
			{
				Name:        "lowercase",
				Type:        ParamTypeBool,
				Description: "Convert to lowercase",
				Default:     false,
				Example:     true,
			},
		},
	})

	// Add other base generators with minimal parameters for now
	basicGenerators := []string{"company", "job_title", "address", "time", "timestamp", "char", "url", "ip", "domain", "username"}
	for _, gen := range basicGenerators {
		pr.RegisterGenerator("base", gen, GeneratorInfo{
			Description: fmt.Sprintf("Generates a random %s", gen),
			Example:     "Generated " + gen,
			Parameters: []ParameterDef{
				{
					Name:        "uppercase",
					Type:        ParamTypeBool,
					Description: "Convert to uppercase",
					Default:     false,
				},
				{
					Name:        "lowercase",
					Type:        ParamTypeBool,
					Description: "Convert to lowercase",
					Default:     false,
				},
				{
					Name:        "prefix",
					Type:        ParamTypeString,
					Description: "Text to add at the beginning",
				},
				{
					Name:        "suffix",
					Type:        ParamTypeString,
					Description: "Text to add at the end",
				},
			},
		})
	}
}

// registerHealthParameters registers parameters for health industry generators
func (pr *ParameterRegistry) registerHealthParameters() {
	healthGenerators := map[string]string{
		"blood_type":        "Generates a random blood type (A+, B-, O+, etc.)",
		"medical_condition": "Generates a random medical condition",
		"medication":        "Generates a random medication name",
		"symptom":          "Generates a random medical symptom",
		"diagnosis":        "Generates a random medical diagnosis",
		"allergy":          "Generates a random allergy",
		"lab_result":       "Generates random lab result values",
		"vital_sign":       "Generates random vital signs",
		"medical_record":   "Generates a complete medical record",
	}

	for gen, desc := range healthGenerators {
		pr.RegisterGenerator("health", gen, GeneratorInfo{
			Description: desc,
			Example:     "Generated " + gen,
			Parameters: []ParameterDef{
				{
					Name:        "format",
					Type:        ParamTypeSelect,
					Description: "Output format",
					Default:     "text",
					Options:     []string{"text", "json"},
				},
			},
		})
	}
}

// registerAviationParameters registers parameters for aviation industry generators
func (pr *ParameterRegistry) registerAviationParameters() {
	aviationGenerators := map[string]string{
		"aircraft_type":   "Generates a random aircraft type (Boeing 737, Airbus A320, etc.)",
		"flight_number":   "Generates a random flight number (AA1234, UA5678, etc.)",
		"airport_code":    "Generates a random airport code (JFK, LAX, etc.)",
		"gate_number":     "Generates a random gate number (A12, B34, etc.)",
		"seat_number":     "Generates a random seat number (12A, 25F, etc.)",
		"flight_status":   "Generates a random flight status (On Time, Delayed, etc.)",
		"baggage_claim":   "Generates a random baggage claim number",
		"flight_schedule": "Generates flight schedule information",
		"flight_info":     "Generates complete flight information",
	}

	for gen, desc := range aviationGenerators {
		pr.RegisterGenerator("aviation", gen, GeneratorInfo{
			Description: desc,
			Example:     "Generated " + gen,
			Parameters: []ParameterDef{
				{
					Name:        "format",
					Type:        ParamTypeSelect,
					Description: "Output format",
					Default:     "text",
					Options:     []string{"text", "json"},
				},
			},
		})
	}
}