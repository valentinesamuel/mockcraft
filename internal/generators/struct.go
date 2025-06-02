package generators

import (
	"reflect"
)

// StructGenerator generates values for struct fields
type StructGenerator struct {
	fieldFuncs map[string]func() interface{}
}

// NewStructGenerator creates a new StructGenerator
func NewStructGenerator() *StructGenerator {
	return &StructGenerator{
		fieldFuncs: make(map[string]func() interface{}),
	}
}

// AddFieldFunc adds a custom function for generating a field value
func (g *StructGenerator) AddFieldFunc(fieldName string, fn func() interface{}) {
	g.fieldFuncs[fieldName] = fn
}

// Generate generates values for a struct
func (g *StructGenerator) Generate(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return nil
	}

	val = val.Elem()
	typ := val.Type()

	// Check if struct implements Fakeable
	if fakeable, ok := v.(interface{ Fake() error }); ok {
		return fakeable.Fake()
	}

	// Generate values for each field
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if fn, ok := g.fieldFuncs[fieldType.Name]; ok {
			field.Set(reflect.ValueOf(fn()))
		}
	}

	return nil
}
