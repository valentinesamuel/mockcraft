package generators

import (
	"fmt"
	"reflect"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

// Generator handles fake data generation
type Generator struct {
	faker *gofakeit.Faker
}

// New creates a new generator instance
func New() *Generator {
	return &Generator{
		faker: gofakeit.New(0), // Use 0 for random seed
	}
}

// GenerateSingle generates a single fake value of the specified type
func (g *Generator) GenerateSingle(dataType string) (interface{}, error) {
	switch dataType {
	case "firstname":
		return g.faker.FirstName(), nil
	case "lastname":
		return g.faker.LastName(), nil
	case "email":
		return g.faker.Email(), nil
	case "phone":
		return g.faker.Phone(), nil
	case "address":
		return g.faker.Address().Address, nil
	case "city":
		return g.faker.Address().City, nil
	case "country":
		return g.faker.Address().Country, nil
	case "company":
		return g.faker.Company(), nil
	case "job":
		return g.faker.JobTitle(), nil
	case "username":
		return g.faker.Username(), nil
	case "password":
		return g.faker.Password(true, true, true, true, false, 12), nil
	case "url":
		return g.faker.URL(), nil
	case "ipv4":
		return g.faker.IPv4Address(), nil
	case "ipv6":
		return g.faker.IPv6Address(), nil
	case "mac":
		return g.faker.MacAddress(), nil
	case "uuid":
		return g.faker.UUID(), nil
	case "creditcard":
		return g.faker.CreditCardNumber(&gofakeit.CreditCardOptions{}), nil
	case "ssn":
		return g.faker.SSN(), nil
	case "date":
		return g.faker.Date(), nil
	case "time":
		return time.Now().Format("15:04:05"), nil
	case "datetime":
		return g.faker.Date(), nil
	case "color":
		return g.faker.Color(), nil
	case "hexcolor":
		return g.faker.HexColor(), nil
	case "rgbcolor":
		return g.faker.RGBColor(), nil
	case "word":
		return g.faker.Word(), nil
	case "sentence":
		return g.faker.Sentence(5), nil
	case "paragraph":
		return g.faker.Paragraph(1, 2, 5, " "), nil
	case "boolean":
		return g.faker.Bool(), nil
	case "int":
		return g.faker.Int64(), nil
	case "float":
		return g.faker.Float64(), nil
	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}
}

// ListGenerators returns a list of all available generator types
func (g *Generator) ListGenerators() []string {
	return []string{
		"firstname", "lastname", "email", "phone", "address", "city", "country",
		"company", "job", "username", "password", "url", "ipv4", "ipv6", "mac",
		"uuid", "creditcard", "ssn", "date", "time", "datetime", "color",
		"hexcolor", "rgbcolor", "word", "sentence", "paragraph", "boolean",
		"int", "float",
	}
}

// GenerateBatch generates multiple fake values of the specified type
func (g *Generator) GenerateBatch(dataType string, count int) ([]interface{}, error) {
	values := make([]interface{}, count)
	for i := 0; i < count; i++ {
		value, err := g.GenerateSingle(dataType)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// GetType returns the Go type for a given data type
func (g *Generator) GetType(dataType string) reflect.Type {
	value, _ := g.GenerateSingle(dataType)
	return reflect.TypeOf(value)
}
