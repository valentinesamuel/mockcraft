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
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Columns     []Column          `json:"columns" yaml:"columns"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Column represents a single column in the schema
type Column struct {
	Name        string            `json:"name" yaml:"name"`
	Type        string            `json:"type" yaml:"type"`
	Industry    string            `json:"industry" yaml:"industry"` // Industry field to identify which generator to use
	Constraints map[string]any    `json:"constraints,omitempty" yaml:"constraints,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// LoadSchema loads a schema from a YAML or JSON file
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
	ext := filepath.Ext(filePath)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &schema); err != nil {
			return nil, fmt.Errorf("failed to parse YAML schema file: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &schema); err != nil {
			return nil, fmt.Errorf("failed to parse JSON schema file: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	return &schema, nil
}

// Validate checks if the schema is valid
func (s *Schema) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("schema name is required")
	}

	if len(s.Columns) == 0 {
		return fmt.Errorf("schema must have at least one column")
	}

	for i, col := range s.Columns {
		if col.Name == "" {
			return fmt.Errorf("column %d: name is required", i+1)
		}
		if col.Type == "" {
			return fmt.Errorf("column %d: type is required", i+1)
		}
		if col.Industry == "" {
			return fmt.Errorf("column %d: industry is required", i+1)
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
