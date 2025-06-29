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
	e.registerPostgreSQLGenerators()
	e.registerSQLiteGenerators()
	e.registerMySQLGenerators()
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
	e.registry.RegisterGenerator("base", "firstname", e.wrapWithParams(e.generateFirstNameWithParams))
	e.registry.RegisterGenerator("base", "lastname", e.wrapWithParams(e.generateLastNameWithParams))
	e.registry.RegisterGenerator("base", "email", e.wrapWithParams(e.generateEmailWithParams))
	e.registry.RegisterGenerator("base", "phone", e.wrapWithParams(e.generatePhoneWithParams))
	e.registry.RegisterGenerator("base", "address", e.wrapWithParams(e.generateAddressWithParams))
	e.registry.RegisterGenerator("base", "company", e.wrapWithParams(e.generateCompanyWithParams))
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
	e.registry.RegisterGenerator("base", "word", e.wrapWithParams(e.generateWordWithParams))
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

// registerPostgreSQLGenerators registers all PostgreSQL-specific generators
func (e *Engine) registerPostgreSQLGenerators() {
	// JSON and JSONB
	e.registry.RegisterGenerator("base", "json", e.wrapWithParams(e.generateJSON))
	e.registry.RegisterGenerator("base", "jsonb", e.wrapWithParams(e.generateJSONB))
	
	// Network types
	e.registry.RegisterGenerator("base", "inet", e.wrapGofakeit(func() interface{} { return e.generateInet() }))
	e.registry.RegisterGenerator("base", "cidr", e.wrapGofakeit(func() interface{} { return e.generateCIDR() }))
	e.registry.RegisterGenerator("base", "macaddr", e.wrapGofakeit(func() interface{} { return e.generateMACAddr() }))
	
	// Binary and money types
	e.registry.RegisterGenerator("base", "bytea", e.wrapWithParams(e.generateBytea))
	e.registry.RegisterGenerator("base", "money", e.wrapWithParams(e.generateMoney))
	
	// Time types
	e.registry.RegisterGenerator("base", "interval", e.wrapWithParams(e.generateInterval))
	
	// Serial types
	e.registry.RegisterGenerator("base", "serial", e.wrapGofakeit(func() interface{} { return e.generateSerial() }))
	e.registry.RegisterGenerator("base", "bigserial", e.wrapGofakeit(func() interface{} { return e.generateBigSerial() }))
	
	// Geometric types
	e.registry.RegisterGenerator("base", "point", e.wrapGofakeit(func() interface{} { return e.generatePoint() }))
	e.registry.RegisterGenerator("base", "line", e.wrapGofakeit(func() interface{} { return e.generateLine() }))
	e.registry.RegisterGenerator("base", "circle", e.wrapGofakeit(func() interface{} { return e.generateCircle() }))
	e.registry.RegisterGenerator("base", "polygon", e.wrapGofakeit(func() interface{} { return e.generatePolygon() }))
	e.registry.RegisterGenerator("base", "path", e.wrapGofakeit(func() interface{} { return e.generatePath() }))
	
	// Full-text search
	e.registry.RegisterGenerator("base", "tsvector", e.wrapGofakeit(func() interface{} { return e.generateTSVector() }))
	e.registry.RegisterGenerator("base", "tsquery", e.wrapGofakeit(func() interface{} { return e.generateTSQuery() }))
	
	// Other types
	e.registry.RegisterGenerator("base", "hstore", e.wrapWithParams(e.generateHStore))
	e.registry.RegisterGenerator("base", "xml", e.wrapGofakeit(func() interface{} { return e.generateXML() }))
	e.registry.RegisterGenerator("base", "pg_array", e.wrapWithParams(e.generateArray))
	
	// Range types
	e.registry.RegisterGenerator("base", "int4range", e.wrapGofakeit(func() interface{} { return e.generateIntRange() }))
	e.registry.RegisterGenerator("base", "int8range", e.wrapGofakeit(func() interface{} { return e.generateInt8Range() }))
	e.registry.RegisterGenerator("base", "numrange", e.wrapGofakeit(func() interface{} { return e.generateNumRange() }))
	e.registry.RegisterGenerator("base", "tsrange", e.wrapGofakeit(func() interface{} { return e.generateTSRange() }))
	e.registry.RegisterGenerator("base", "tstzrange", e.wrapGofakeit(func() interface{} { return e.generateTSTZRange() }))
	e.registry.RegisterGenerator("base", "daterange", e.wrapGofakeit(func() interface{} { return e.generateDateRange() }))
	
	// Bit string types
	e.registry.RegisterGenerator("base", "bit", e.wrapWithParams(e.generateBit))
	
	// Additional geometric types
	e.registry.RegisterGenerator("base", "box", e.wrapGofakeit(func() interface{} { return e.generateBox() }))
	e.registry.RegisterGenerator("base", "lseg", e.wrapGofakeit(func() interface{} { return e.generateLseg() }))
}

// registerSQLiteGenerators registers all SQLite-specific generators
func (e *Engine) registerSQLiteGenerators() {
	// SQLite stores most complex types as TEXT, so we reuse PostgreSQL generators
	// but ensure they output SQLite-compatible formats
	
	// JSON (stored as TEXT in SQLite)
	e.registry.RegisterGenerator("base", "sqlite_json", e.wrapWithParams(e.generateSQLiteJSON))
	
	// Arrays (stored as JSON TEXT in SQLite)
	e.registry.RegisterGenerator("base", "sqlite_array", e.wrapWithParams(e.generateSQLiteArray))
	
	// Binary data (BLOB type)
	e.registry.RegisterGenerator("base", "sqlite_blob", e.wrapWithParams(e.generateSQLiteBlob))
	
	// Auto-increment IDs
	e.registry.RegisterGenerator("base", "sqlite_autoincrement", e.wrapGofakeit(func() interface{} { return e.generateSQLiteAutoIncrement() }))
	
	// Datetime as TEXT (ISO8601 format)
	e.registry.RegisterGenerator("base", "sqlite_datetime", e.wrapGofakeit(func() interface{} { return e.generateSQLiteDatetime() }))
	e.registry.RegisterGenerator("base", "sqlite_date", e.wrapGofakeit(func() interface{} { return e.generateSQLiteDate() }))
	e.registry.RegisterGenerator("base", "sqlite_time", e.wrapGofakeit(func() interface{} { return e.generateSQLiteTime() }))
	
	// Boolean as INTEGER (0/1)
	e.registry.RegisterGenerator("base", "sqlite_boolean", e.wrapGofakeit(func() interface{} { return e.generateSQLiteBoolean() }))
	
	// Numeric types
	e.registry.RegisterGenerator("base", "sqlite_real", e.wrapWithParams(e.generateSQLiteReal))
	e.registry.RegisterGenerator("base", "sqlite_decimal", e.wrapWithParams(e.generateSQLiteDecimal))
	
	// UUID as TEXT
	e.registry.RegisterGenerator("base", "sqlite_uuid", e.wrapGofakeit(func() interface{} { return e.generateUUID() }))
	
	// Network types as TEXT
	e.registry.RegisterGenerator("base", "sqlite_inet", e.wrapGofakeit(func() interface{} { return e.generateInet() }))
	e.registry.RegisterGenerator("base", "sqlite_macaddr", e.wrapGofakeit(func() interface{} { return e.generateMACAddr() }))
	
	// Geometric types as TEXT (JSON or custom format)
	e.registry.RegisterGenerator("base", "sqlite_point", e.wrapGofakeit(func() interface{} { return e.generateSQLitePoint() }))
	e.registry.RegisterGenerator("base", "sqlite_polygon", e.wrapGofakeit(func() interface{} { return e.generateSQLitePolygon() }))
}

// registerMySQLGenerators registers all MySQL-specific generators
func (e *Engine) registerMySQLGenerators() {
	// JSON type (MySQL 5.7+)
	e.registry.RegisterGenerator("base", "mysql_json", e.wrapWithParams(e.generateMySQLJSON))
	
	// Geometric types (MySQL spatial extensions)
	e.registry.RegisterGenerator("base", "mysql_point", e.wrapGofakeit(func() interface{} { return e.generateMySQLPoint() }))
	e.registry.RegisterGenerator("base", "mysql_linestring", e.wrapGofakeit(func() interface{} { return e.generateMySQLLineString() }))
	e.registry.RegisterGenerator("base", "mysql_polygon", e.wrapGofakeit(func() interface{} { return e.generateMySQLPolygon() }))
	e.registry.RegisterGenerator("base", "mysql_multipoint", e.wrapGofakeit(func() interface{} { return e.generateMySQLMultiPoint() }))
	e.registry.RegisterGenerator("base", "mysql_multilinestring", e.wrapGofakeit(func() interface{} { return e.generateMySQLMultiLineString() }))
	e.registry.RegisterGenerator("base", "mysql_multipolygon", e.wrapGofakeit(func() interface{} { return e.generateMySQLMultiPolygon() }))
	e.registry.RegisterGenerator("base", "mysql_geometry", e.wrapGofakeit(func() interface{} { return e.generateMySQLGeometry() }))
	e.registry.RegisterGenerator("base", "mysql_geometrycollection", e.wrapGofakeit(func() interface{} { return e.generateMySQLGeometryCollection() }))
	
	// Binary types
	e.registry.RegisterGenerator("base", "mysql_binary", e.wrapWithParams(e.generateMySQLBinary))
	e.registry.RegisterGenerator("base", "mysql_varbinary", e.wrapWithParams(e.generateMySQLVarBinary))
	e.registry.RegisterGenerator("base", "mysql_blob", e.wrapWithParams(e.generateMySQLBlob))
	e.registry.RegisterGenerator("base", "mysql_tinyblob", e.wrapGofakeit(func() interface{} { return e.generateMySQLTinyBlob() }))
	e.registry.RegisterGenerator("base", "mysql_mediumblob", e.wrapWithParams(e.generateMySQLMediumBlob))
	e.registry.RegisterGenerator("base", "mysql_longblob", e.wrapWithParams(e.generateMySQLLongBlob))
	
	// Text types
	e.registry.RegisterGenerator("base", "mysql_tinytext", e.wrapGofakeit(func() interface{} { return e.generateMySQLTinyText() }))
	e.registry.RegisterGenerator("base", "mysql_mediumtext", e.wrapWithParams(e.generateMySQLMediumText))
	e.registry.RegisterGenerator("base", "mysql_longtext", e.wrapWithParams(e.generateMySQLLongText))
	
	// Integer types
	e.registry.RegisterGenerator("base", "mysql_tinyint", e.wrapWithParams(e.generateMySQLTinyInt))
	e.registry.RegisterGenerator("base", "mysql_smallint", e.wrapWithParams(e.generateMySQLSmallInt))
	e.registry.RegisterGenerator("base", "mysql_mediumint", e.wrapWithParams(e.generateMySQLMediumInt))
	e.registry.RegisterGenerator("base", "mysql_bigint", e.wrapWithParams(e.generateMySQLBigInt))
	
	// Decimal types
	e.registry.RegisterGenerator("base", "mysql_decimal", e.wrapWithParams(e.generateMySQLDecimal))
	e.registry.RegisterGenerator("base", "mysql_numeric", e.wrapWithParams(e.generateMySQLNumeric))
	e.registry.RegisterGenerator("base", "mysql_double", e.wrapWithParams(e.generateMySQLDouble))
	e.registry.RegisterGenerator("base", "mysql_real", e.wrapWithParams(e.generateMySQLReal))
	
	// Date and time types
	e.registry.RegisterGenerator("base", "mysql_datetime", e.wrapGofakeit(func() interface{} { return e.generateMySQLDateTime() }))
	e.registry.RegisterGenerator("base", "mysql_timestamp", e.wrapGofakeit(func() interface{} { return e.generateMySQLTimestamp() }))
	e.registry.RegisterGenerator("base", "mysql_time", e.wrapGofakeit(func() interface{} { return e.generateMySQLTime() }))
	e.registry.RegisterGenerator("base", "mysql_year", e.wrapGofakeit(func() interface{} { return e.generateMySQLYear() }))
	
	// Bit type
	e.registry.RegisterGenerator("base", "mysql_bit", e.wrapWithParams(e.generateMySQLBit))
	
	// Enum and Set types
	e.registry.RegisterGenerator("base", "mysql_enum", e.wrapWithParams(e.generateMySQLEnum))
	e.registry.RegisterGenerator("base", "mysql_set", e.wrapWithParams(e.generateMySQLSet))
	
	// Auto-increment types
	e.registry.RegisterGenerator("base", "mysql_auto_increment", e.wrapGofakeit(func() interface{} { return e.generateMySQLAutoIncrement() }))
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