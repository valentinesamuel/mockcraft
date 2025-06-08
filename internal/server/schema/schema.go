package schema

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// Schema represents the structure of a data generation schema
type Schema struct {
	Tables []Table `yaml:"tables"`
}

// Table represents a table in the schema
type Table struct {
	Name    string   `yaml:"name"`
	Count   int      `yaml:"count"`
	Columns []Column `yaml:"columns"`
}

// Column represents a column in a table
type Column struct {
    Name      string                 `yaml:"name"`
    Type      string                 `yaml:"type"`
    Generator string                 `yaml:"generator,omitempty"`
    Params    map[string]interface{} `yaml:"params,omitempty"`
    Nullable  bool                   `yaml:"nullable,omitempty"`
    Unique    bool                   `yaml:"unique,omitempty"`
}

// Parse parses a YAML schema file
func Parse(filePath string) (*Schema, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open schema file: %w", err)
	}
	defer file.Close()

	return ParseReader(file)
}

// ParseReader parses a YAML schema from a reader
func ParseReader(reader io.Reader) (*Schema, error) {
	var schema Schema
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode schema: %w", err)
	}

	if err := validateSchema(&schema); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	return &schema, nil
}

func validateSchema(schema *Schema) error {
    if len(schema.Tables) == 0 {
        return fmt.Errorf("schema must contain at least one table")
    }

    for i, table := range schema.Tables {
        if table.Name == "" {
            return fmt.Errorf("table %d: name is required", i)
        }

        if table.Count <= 0 {
            return fmt.Errorf("table %s: count must be greater than 0", table.Name)
        }

        if len(table.Columns) == 0 {
            return fmt.Errorf("table %s: must contain at least one column", table.Name)
        }

        columnNames := make(map[string]bool)
        for j, column := range table.Columns {
            if column.Name == "" {
                return fmt.Errorf("table %s, column %d: name is required", table.Name, j)
            }

            if column.Type == "" {
                return fmt.Errorf("table %s, column %s: type is required", table.Name, column.Name)
            }

            if columnNames[column.Name] {
                return fmt.Errorf("table %s: duplicate column name: %s", table.Name, column.Name)
            }
            columnNames[column.Name] = true

            // Optional: Validate generator exists if specified
            if column.Generator != "" {
                // You could add registry validation here if needed
                // This would require importing the registry package
            }
        }
    }

    return nil
}
