package generators

import (
	"fmt"
	"math/rand"
	"reflect"
)

// Fakeable is the interface that types must implement to control their own generation
type Fakeable interface {
	// Fake should populate the struct with random data
	Fake() error
}

// FieldFunc is a function that generates random data for a specific field
type FieldFunc func() interface{}

// FieldMap maps struct field names to their custom generation functions
type FieldMap map[string]FieldFunc

// Generator defines the interface that all generators must implement
type Generator interface {
	// GenerateByType generates data based on type string and parameters
	GenerateByType(dataType string, params map[string]interface{}) (interface{}, error)
	// GetAvailableTypes returns all supported types for this generator
	GetAvailableTypes() []string
}

// StructGenerator provides functionality to generate random data for structs
type StructGenerator struct {
	fieldFuncs FieldMap
}

// NewStructGenerator creates a new struct generator
func NewStructGenerator() *StructGenerator {
	return &StructGenerator{
		fieldFuncs: make(FieldMap),
	}
}

// AddFieldFunc adds a custom generation function for a field
func (sg *StructGenerator) AddFieldFunc(fieldName string, fn FieldFunc) {
	sg.fieldFuncs[fieldName] = fn
}

// Generate generates random data for a struct
func (sg *StructGenerator) Generate(v interface{}) error {
	// If the type implements Fakeable, use its Fake method
	if fakeable, ok := v.(Fakeable); ok {
		return fakeable.Fake()
	}

	// Otherwise, use reflection to generate data for each field
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct, got %v", val.Kind())
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Check if we have a custom function for this field
		if fn, ok := sg.fieldFuncs[fieldType.Name]; ok {
			field.Set(reflect.ValueOf(fn()))
			continue
		}

		// Try to generate data based on field type
		if err := sg.generateField(field, fieldType); err != nil {
			return fmt.Errorf("failed to generate field %s: %w", fieldType.Name, err)
		}
	}

	return nil
}

// generateField generates random data for a field based on its type
func (sg *StructGenerator) generateField(field reflect.Value, fieldType reflect.StructField) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(generateString(fieldType))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		field.SetInt(generateInt(fieldType))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field.SetUint(generateUint(fieldType))
	case reflect.Float32, reflect.Float64:
		field.SetFloat(generateFloat(fieldType))
	case reflect.Bool:
		field.SetBool(generateBool(fieldType))
	case reflect.Struct:
		if err := sg.Generate(field.Addr().Interface()); err != nil {
			return err
		}
	case reflect.Slice:
		if err := sg.generateSlice(field, fieldType); err != nil {
			return err
		}
	case reflect.Map:
		if err := sg.generateMap(field, fieldType); err != nil {
			return err
		}
	}
	return nil
}

// Helper functions for generating random data
func generateString(fieldType reflect.StructField) string {
	// TODO: Implement based on field tags or type
	return "random string"
}

func generateInt(fieldType reflect.StructField) int64 {
	// TODO: Implement based on field tags or type
	return rand.Int63()
}

func generateUint(fieldType reflect.StructField) uint64 {
	// TODO: Implement based on field tags or type
	return rand.Uint64()
}

func generateFloat(fieldType reflect.StructField) float64 {
	// TODO: Implement based on field tags or type
	return rand.Float64()
}

func generateBool(fieldType reflect.StructField) bool {
	// TODO: Implement based on field tags or type
	return rand.Float32() < 0.5
}

func (sg *StructGenerator) generateSlice(field reflect.Value, fieldType reflect.StructField) error {
	// TODO: Implement slice generation
	return nil
}

func (sg *StructGenerator) generateMap(field reflect.Value, fieldType reflect.StructField) error {
	// TODO: Implement map generation
	return nil
}

// Example usage:
/*
type User struct {
	Name     string    `fake:"{firstname}"`
	Email    string    `fake:"{email}"`
	Age      int       `fake:"{number:18,65}"`
	Birthday time.Time `fake:"{date}"`
	Active   bool      `fake:"{bool}"`
}

// Custom generation for a field
func (u *User) Fake() error {
	u.Name = "Custom Name"
	u.Email = "custom@email.com"
	u.Age = 25
	u.Birthday = time.Now()
	u.Active = true
	return nil
}

// Using the generator
func Example() {
	// Create a new generator
	gen := NewStructGenerator()

	// Add custom field function
	gen.AddFieldFunc("Name", func() interface{} {
		return "John Doe"
	})

	// Generate data for a struct
	user := &User{}
	if err := gen.Generate(user); err != nil {
		panic(err)
	}
}
*/
