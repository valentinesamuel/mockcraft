package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Schema represents a database seeding schema
type Schema struct {
	Tables []Table `yaml:"tables"`
}

// Table represents a database table definition
type Table struct {
	Name        string       `yaml:"name"`
	Count       int          `yaml:"count"`
	Columns     []Column     `yaml:"columns"`
	Relations   []Relation   `yaml:"relations"`
	Constraints []Constraint `yaml:"constraints"`
}

// Column represents a table column definition
type Column struct {
	Name       string                 `yaml:"name"`
	Type       string                 `yaml:"type"`
	Generator  string                 `yaml:"generator"`
	Params     map[string]interface{} `yaml:"params"`
	Nullable   bool                   `yaml:"nullable"`
	Unique     bool                   `yaml:"unique"`
	PrimaryKey bool                   `yaml:"primary_key"`
}

// Relation represents a relationship between tables
type Relation struct {
	Type       string `yaml:"type"` // one-to-one, one-to-many, many-to-many
	FromTable  string `yaml:"from_table"`
	FromColumn string `yaml:"from_column"`
	ToTable    string `yaml:"to_table"`
	ToColumn   string `yaml:"to_column"`
}

// Constraint represents a table constraint
type Constraint struct {
	Type      string   `yaml:"type"` // unique, check, foreign_key
	Columns   []string `yaml:"columns"`
	Condition string   `yaml:"condition,omitempty"`
}

// ValidateSchema validates the schema definition
func ValidateSchema(schema *Schema) error {
	// Check for empty schema
	if len(schema.Tables) == 0 {
		return fmt.Errorf("schema must contain at least one table")
	}

	// Validate each table
	for _, table := range schema.Tables {
		if err := validateTable(&table); err != nil {
			return fmt.Errorf("invalid table '%s': %v", table.Name, err)
		}
	}

	// Check for circular dependencies
	if err := checkCircularDependencies(schema); err != nil {
		return fmt.Errorf("circular dependency detected: %v", err)
	}

	return nil
}

// validateTable validates a single table definition
func validateTable(table *Table) error {
	// Check table name
	if table.Name == "" {
		return fmt.Errorf("table name is required")
	}

	// Check count
	if table.Count < 0 {
		return fmt.Errorf("count must be non-negative")
	}

	// Check columns
	if len(table.Columns) == 0 {
		return fmt.Errorf("table must have at least one column")
	}

	// Validate each column
	for _, column := range table.Columns {
		if err := validateColumn(&column); err != nil {
			return fmt.Errorf("invalid column '%s': %v", column.Name, err)
		}
	}

	// Validate relations
	for _, relation := range table.Relations {
		if err := validateRelation(&relation); err != nil {
			return fmt.Errorf("invalid relation: %v", err)
		}
	}

	return nil
}

// validateColumn validates a single column definition
func validateColumn(column *Column) error {
	// Check column name
	if column.Name == "" {
		return fmt.Errorf("column name is required")
	}

	// Check type
	if column.Type == "" {
		return fmt.Errorf("column type is required")
	}

	// Check generator
	if column.Generator == "" {
		return fmt.Errorf("generator is required")
	}

	return nil
}

// validateRelation validates a single relation definition
func validateRelation(relation *Relation) error {
	// Check relation type
	switch relation.Type {
	case "one-to-one", "one-to-many", "many-to-many":
		// Valid types
	default:
		return fmt.Errorf("invalid relation type: %s", relation.Type)
	}

	// Check required fields
	if relation.FromTable == "" {
		return fmt.Errorf("from_table is required")
	}
	if relation.FromColumn == "" {
		return fmt.Errorf("from_column is required")
	}
	if relation.ToTable == "" {
		return fmt.Errorf("to_table is required")
	}
	if relation.ToColumn == "" {
		return fmt.Errorf("to_column is required")
	}

	return nil
}

// checkCircularDependencies checks for circular dependencies between tables
func checkCircularDependencies(schema *Schema) error {
	// Create adjacency list for tables
	graph := make(map[string][]string)
	for _, table := range schema.Tables {
		graph[table.Name] = []string{}
		for _, relation := range table.Relations {
			graph[table.Name] = append(graph[table.Name], relation.ToTable)
		}
	}

	// Check for cycles using DFS
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(string) error
	dfs = func(table string) error {
		visited[table] = true
		recStack[table] = true

		for _, neighbor := range graph[table] {
			if !visited[neighbor] {
				if err := dfs(neighbor); err != nil {
					return err
				}
			} else if recStack[neighbor] {
				return fmt.Errorf("circular dependency detected between tables")
			}
		}

		recStack[table] = false
		return nil
	}

	// Run DFS for each table
	for table := range graph {
		if !visited[table] {
			if err := dfs(table); err != nil {
				return err
			}
		}
	}

	return nil
}

// LoadSchema loads and validates a schema from a YAML file
func LoadSchema(path string) (*Schema, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading schema file: %v", err)
	}

	// Parse YAML
	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("error parsing schema file: %v", err)
	}

	// Validate schema
	if err := ValidateSchema(&schema); err != nil {
		return nil, fmt.Errorf("invalid schema: %v", err)
	}

	return &schema, nil
}
