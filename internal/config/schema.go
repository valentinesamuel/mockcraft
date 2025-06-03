package config

import (
	"fmt"
	"os"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"gopkg.in/yaml.v3"
)

// LoadSchema loads a schema from a YAML file
func LoadSchema(path string) (*types.Schema, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	// Parse YAML
	var schema types.Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema file: %w", err)
	}

	// Validate schema
	if err := validateSchema(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

// validateSchema validates the schema
func validateSchema(schema *types.Schema) error {
	// Validate tables
	if len(schema.Tables) == 0 {
		return fmt.Errorf("schema must contain at least one table")
	}

	// Validate each table
	for i, table := range schema.Tables {
		// Validate table name
		if table.Name == "" {
			return fmt.Errorf("table %d: name is required", i)
		}

		// Validate columns
		if len(table.Columns) == 0 {
			return fmt.Errorf("table %s: must contain at least one column", table.Name)
		}

		// Validate each column
		for j, col := range table.Columns {
			// Validate column name
			if col.Name == "" {
				return fmt.Errorf("table %s, column %d: name is required", table.Name, j)
			}

			// Validate column type
			if col.Type == "" {
				return fmt.Errorf("table %s, column %s: type is required", table.Name, col.Name)
			}
		}

		// Validate indexes
		for j, index := range table.Indexes {
			// Validate index name
			if index.Name == "" {
				return fmt.Errorf("table %s, index %d: name is required", table.Name, j)
			}

			// Validate index columns
			if len(index.Columns) == 0 {
				return fmt.Errorf("table %s, index %s: must contain at least one column", table.Name, index.Name)
			}

			// Validate index type
			if index.Type == "" {
				return fmt.Errorf("table %s, index %s: type is required", table.Name, index.Name)
			}
		}
	}

	return nil
}
