package generators

import (
	"fmt"
	"math/rand"

	"github.com/valentinesamuel/mockcraft/internal/generators/registry"
)

// Engine is the unified generator engine that provides a single interface
// for all generator operations across CLI, API, and seeder modules
type Engine struct {
	registry      *registry.IndustryRegistry
	paramRegistry *ParameterRegistry
	rand          *rand.Rand
}

// NewEngine creates a new unified generator engine
func NewEngine() *Engine {
	engine := &Engine{
		registry:      registry.NewIndustryRegistry(),
		paramRegistry: NewParameterRegistry(),
		rand:          rand.New(rand.NewSource(rand.Int63())),
	}
	
	// Register all generators from all industries
	engine.registerAllGenerators()
	
	return engine
}

// registerAllGenerators registers all generators from all industries
func (e *Engine) registerAllGenerators() {
	e.registerBaseGenerators()
	e.registerMongoDBGenerators()
	e.registerAviationGenerators()
	e.registerHealthGenerators()
}

// SetSeed sets the random seed for reproducible results
func (e *Engine) SetSeed(seed int64) {
	e.rand.Seed(seed)
}

// Generate generates data using industry and generator name with optional parameters
func (e *Engine) Generate(industry, generator string, params map[string]interface{}) (interface{}, error) {
	generatorFunc, err := e.registry.GetGenerator(industry, generator)
	if err != nil {
		return nil, fmt.Errorf("failed to get generator %s from industry %s: %w", generator, industry, err)
	}

	// Convert and validate parameters
	validatedParams, err := e.paramRegistry.ConvertAndValidateParams(industry, generator, params)
	if err != nil {
		return nil, fmt.Errorf("parameter validation failed: %w", err)
	}

	// Add random source to params if not provided
	if validatedParams == nil {
		validatedParams = make(map[string]interface{})
	}
	validatedParams["_rand"] = e.rand

	// Generate the value
	value, err := generatorFunc(validatedParams)
	if err != nil {
		return nil, err
	}

	// Apply text transformations (uppercase, lowercase, prefix, suffix, etc.)
	transformedValue := ApplyTextTransformations(value, validatedParams)

	return transformedValue, nil
}

// ValidateGenerator checks if a generator exists for the given industry
func (e *Engine) ValidateGenerator(industry, generator string) error {
	if !e.registry.HasGenerator(industry, generator) {
		return fmt.Errorf("generator %s not found for industry %s", generator, industry)
	}
	return nil
}

// ListIndustries returns all available industries
func (e *Engine) ListIndustries() []string {
	return e.registry.ListIndustries()
}

// ListGenerators returns all generators for a given industry
func (e *Engine) ListGenerators(industry string) ([]string, error) {
	return e.registry.ListGenerators(industry)
}

// GetAllGenerators returns a map of all industries and their generators
func (e *Engine) GetAllGenerators() map[string][]string {
	result := make(map[string][]string)
	industries := e.ListIndustries()
	
	for _, industry := range industries {
		generators, err := e.ListGenerators(industry)
		if err == nil {
			result[industry] = generators
		}
	}
	
	return result
}

// GenerateRow generates a complete row of data based on column specifications
// This is used by the seeder to generate database rows
func (e *Engine) GenerateRow(columns []ColumnSpec, existingData map[string]interface{}) (map[string]interface{}, error) {
	row := make(map[string]interface{})
	
	// Copy existing data (for foreign keys, etc.)
	for k, v := range existingData {
		row[k] = v
	}
	
	for _, col := range columns {
		// Skip if value already exists (e.g., from foreign key relationships)
		if _, exists := row[col.Name]; exists {
			continue
		}
		
		value, err := e.Generate(col.Industry, col.Generator, col.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to generate value for column %s: %w", col.Name, err)
		}
		
		row[col.Name] = value
	}
	
	return row, nil
}

// ColumnSpec represents a column specification for row generation
type ColumnSpec struct {
	Name      string
	Industry  string
	Generator string
	Params    map[string]interface{}
}

// Global engine instance
var globalEngine *Engine

// GetGlobalEngine returns the global engine instance
func GetGlobalEngine() *Engine {
	if globalEngine == nil {
		globalEngine = NewEngine()
	}
	return globalEngine
}

// SetGlobalSeed sets the seed for the global engine
func SetGlobalSeed(seed int64) {
	GetGlobalEngine().SetSeed(seed)
}

// GetGeneratorInfo returns parameter information for a specific generator
func (e *Engine) GetGeneratorInfo(industry, generator string) (*GeneratorInfo, error) {
	return e.paramRegistry.GetGeneratorInfo(industry, generator)
}

// GetAllGeneratorInfo returns all generator information with parameters
func (e *Engine) GetAllGeneratorInfo() map[string]map[string]GeneratorInfo {
	return e.paramRegistry.GetAllGeneratorInfo()
}

// registerBaseGenerators registers all base generators (gofakeit wrappers)
func (e *Engine) registerBaseGenerators() {
	// Import base generator functions
	e.registry.RegisterGenerator("base", "uuid", e.wrapGofakeit(func() interface{} { return e.generateUUID() }))
	e.registry.RegisterGenerator("base", "firstname", e.wrapGofakeit(func() interface{} { return e.generateFirstName() }))
	e.registry.RegisterGenerator("base", "lastname", e.wrapGofakeit(func() interface{} { return e.generateLastName() }))
	e.registry.RegisterGenerator("base", "email", e.wrapGofakeit(func() interface{} { return e.generateEmail() }))
	e.registry.RegisterGenerator("base", "phone", e.wrapGofakeit(func() interface{} { return e.generatePhone() }))
	e.registry.RegisterGenerator("base", "address", e.wrapGofakeit(func() interface{} { return e.generateAddress() }))
	e.registry.RegisterGenerator("base", "company", e.wrapGofakeit(func() interface{} { return e.generateCompany() }))
	e.registry.RegisterGenerator("base", "job_title", e.wrapGofakeit(func() interface{} { return e.generateJobTitle() }))
	e.registry.RegisterGenerator("base", "date", e.wrapGofakeit(func() interface{} { return e.generateDate() }))
	e.registry.RegisterGenerator("base", "datetime", e.wrapGofakeit(func() interface{} { return e.generateDateTime() }))
	e.registry.RegisterGenerator("base", "time", e.wrapGofakeit(func() interface{} { return e.generateTime() }))
	e.registry.RegisterGenerator("base", "timestamp", e.wrapGofakeit(func() interface{} { return e.generateTimestamp() }))
	e.registry.RegisterGenerator("base", "number", e.wrapWithParams(e.generateNumber))
	e.registry.RegisterGenerator("base", "float", e.wrapWithParams(e.generateFloat))
	e.registry.RegisterGenerator("base", "boolean", e.wrapGofakeit(func() interface{} { return e.generateBoolean() }))
	e.registry.RegisterGenerator("base", "text", e.wrapWithParams(e.generateText))
	e.registry.RegisterGenerator("base", "paragraph", e.wrapWithParams(e.generateParagraph))
	e.registry.RegisterGenerator("base", "sentence", e.wrapGofakeit(func() interface{} { return e.generateSentence() }))
	e.registry.RegisterGenerator("base", "word", e.wrapGofakeit(func() interface{} { return e.generateWord() }))
	e.registry.RegisterGenerator("base", "char", e.wrapGofakeit(func() interface{} { return e.generateChar() }))
	e.registry.RegisterGenerator("base", "url", e.wrapGofakeit(func() interface{} { return e.generateURL() }))
	e.registry.RegisterGenerator("base", "ip", e.wrapGofakeit(func() interface{} { return e.generateIP() }))
	e.registry.RegisterGenerator("base", "domain", e.wrapGofakeit(func() interface{} { return e.generateDomain() }))
	e.registry.RegisterGenerator("base", "username", e.wrapGofakeit(func() interface{} { return e.generateUsername() }))
	e.registry.RegisterGenerator("base", "password", e.wrapWithParams(e.generatePassword))
	e.registry.RegisterGenerator("base", "enum", e.wrapWithParams(e.generateEnum))
	e.registry.RegisterGenerator("base", "random_int", e.wrapWithParams(e.generateRandomInt))
	e.registry.RegisterGenerator("base", "null", e.wrapGofakeit(func() interface{} { return e.generateNull() }))
	e.registry.RegisterGenerator("base", "embedded_document", e.wrapWithParams(e.generateEmbeddedDocument))
	e.registry.RegisterGenerator("base", "array_of_strings", e.wrapWithParams(e.generateArrayOfStrings))
}

// registerMongoDBGenerators registers all MongoDB-specific generators
func (e *Engine) registerMongoDBGenerators() {
	e.registry.RegisterGenerator("base", "mongodb_objectid", e.wrapGofakeit(func() interface{} { return e.generateMongoDBObjectID() }))
	e.registry.RegisterGenerator("base", "mongodb_decimal128", e.wrapWithParams(e.generateMongoDBDecimal128))
	e.registry.RegisterGenerator("base", "mongodb_binary", e.wrapWithParams(e.generateMongoDBBinary))
	e.registry.RegisterGenerator("base", "mongodb_timestamp", e.wrapGofakeit(func() interface{} { return e.generateMongoDBTimestamp() }))
	e.registry.RegisterGenerator("base", "mongodb_regex", e.wrapGofakeit(func() interface{} { return e.generateMongoDBRegex() }))
	e.registry.RegisterGenerator("base", "mongodb_javascript", e.wrapGofakeit(func() interface{} { return e.generateMongoDBJavaScript() }))
	e.registry.RegisterGenerator("base", "mongodb_minkey", e.wrapGofakeit(func() interface{} { return e.generateMongoDBMinKey() }))
	e.registry.RegisterGenerator("base", "mongodb_maxkey", e.wrapGofakeit(func() interface{} { return e.generateMongoDBMaxKey() }))
}

// registerAviationGenerators registers all aviation-specific generators
func (e *Engine) registerAviationGenerators() {
	e.registry.RegisterGenerator("aviation", "aircraft_type", e.wrapGofakeit(func() interface{} { return e.generateAircraftType() }))
	e.registry.RegisterGenerator("aviation", "flight_number", e.wrapGofakeit(func() interface{} { return e.generateFlightNumber() }))
	e.registry.RegisterGenerator("aviation", "airport_code", e.wrapGofakeit(func() interface{} { return e.generateAirportCode() }))
	e.registry.RegisterGenerator("aviation", "gate_number", e.wrapGofakeit(func() interface{} { return e.generateGateNumber() }))
	e.registry.RegisterGenerator("aviation", "seat_number", e.wrapGofakeit(func() interface{} { return e.generateSeatNumber() }))
	e.registry.RegisterGenerator("aviation", "flight_status", e.wrapGofakeit(func() interface{} { return e.generateFlightStatus() }))
	e.registry.RegisterGenerator("aviation", "baggage_claim", e.wrapGofakeit(func() interface{} { return e.generateBaggageClaim() }))
	e.registry.RegisterGenerator("aviation", "flight_schedule", e.wrapGofakeit(func() interface{} { return e.generateFlightSchedule() }))
	e.registry.RegisterGenerator("aviation", "flight_info", e.wrapGofakeit(func() interface{} { return e.generateFlightInfo() }))
}

// registerHealthGenerators registers all health-specific generators
func (e *Engine) registerHealthGenerators() {
	e.registry.RegisterGenerator("health", "blood_type", e.wrapGofakeit(func() interface{} { return e.generateBloodType() }))
	e.registry.RegisterGenerator("health", "medical_condition", e.wrapGofakeit(func() interface{} { return e.generateMedicalCondition() }))
	e.registry.RegisterGenerator("health", "medication", e.wrapGofakeit(func() interface{} { return e.generateMedication() }))
	e.registry.RegisterGenerator("health", "symptom", e.wrapGofakeit(func() interface{} { return e.generateSymptom() }))
	e.registry.RegisterGenerator("health", "diagnosis", e.wrapGofakeit(func() interface{} { return e.generateDiagnosis() }))
	e.registry.RegisterGenerator("health", "allergy", e.wrapGofakeit(func() interface{} { return e.generateAllergy() }))
	e.registry.RegisterGenerator("health", "lab_result", e.wrapGofakeit(func() interface{} { return e.generateLabResult() }))
	e.registry.RegisterGenerator("health", "vital_sign", e.wrapGofakeit(func() interface{} { return e.generateVitalSigns() }))
	e.registry.RegisterGenerator("health", "medical_record", e.wrapGofakeit(func() interface{} { return e.generateMedicalRecord() }))
}

// wrapGofakeit wraps a simple generator function
func (e *Engine) wrapGofakeit(fn func() interface{}) registry.GeneratorFunc {
	return func(params map[string]interface{}) (interface{}, error) {
		return fn(), nil
	}
}

// wrapWithParams wraps a generator function that takes parameters
func (e *Engine) wrapWithParams(fn func(map[string]interface{}) (interface{}, error)) registry.GeneratorFunc {
	return fn
}