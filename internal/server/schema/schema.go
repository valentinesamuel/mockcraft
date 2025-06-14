package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Schema represents the structure of a data generation schema
type Schema struct {
	Tables    []Table    `yaml:"tables"`
	Relations []Relation `yaml:"relations"`
}

// Table represents a table in the schema
type Table struct {
	Name      string   `yaml:"name"`
	Count     int      `yaml:"count"`
	Industry  string   `yaml:"industry"`
	Generator string   `yaml:"generator"`
	Columns   []Column `yaml:"columns"`
}

// Column represents a single column in the schema
type Column struct {
	Name       string         `yaml:"name"`
	Type       string         `yaml:"type"`
	Generator  string         `yaml:"generator"`
	Industry   string         `yaml:"industry"`
	IsPrimary  bool           `yaml:"is_primary"`
	IsNullable bool           `yaml:"is_nullable"`
	Params     map[string]any `yaml:"params,omitempty"`
}

// Relation represents a relationship between tables
type Relation struct {
	Type       string `yaml:"type"`
	FromTable  string `yaml:"from_table"`
	FromColumn string `yaml:"from_column"`
	ToTable    string `yaml:"to_table"`
	ToColumn   string `yaml:"to_column"`
}

// LoadSchema loads a schema from a YAML file
func LoadSchema(filePath string) (*Schema, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open schema file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse YAML schema file: %w", err)
	}

	return &schema, nil
}

// Validate checks if the schema is valid
func (s *Schema) Validate() error {
	if len(s.Tables) == 0 {
		return fmt.Errorf("schema must have at least one table")
	}

	for i, table := range s.Tables {
		if table.Name == "" {
			return fmt.Errorf("table %d: name is required", i+1)
		}
		if table.Count <= 0 {
			return fmt.Errorf("table %s: count must be greater than 0", table.Name)
		}
		if len(table.Columns) == 0 {
			return fmt.Errorf("table %s: must have at least one column", table.Name)
		}

		for j, col := range table.Columns {
			if col.Name == "" {
				return fmt.Errorf("table %s, column %d: name is required", table.Name, j+1)
			}
			if col.Type == "" {
				return fmt.Errorf("table %s, column %s: type is required", table.Name, col.Name)
			}
			if col.Generator == "" {
				return fmt.Errorf("table %s, column %s: generator is required", table.Name, col.Name)
			}
			if col.Industry == "" {
				return fmt.Errorf("table %s, column %s: industry is required", table.Name, col.Name)
			}
		}
	}

	return nil
}

// Save saves the schema to a JSON file
func (s *Schema) Save(filePath string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	return nil
}
