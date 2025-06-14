package schema

import (
	"fmt"

	"github.com/valentinesamuel/mockcraft/internal/generators"
)

// Schema represents a database schema
type Schema struct {
	Tables    []Table    `json:"tables"`
	Relations []Relation `json:"relations,omitempty"`
}

// Table represents a database table
type Table struct {
	Name    string   `json:"name"`
	Count   int      `json:"count"`
	Columns []Column `json:"columns"`
	Indexes []Index  `json:"indexes,omitempty"`
}

// Column represents a database column
type Column struct {
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Industry  string                 `json:"industry,omitempty"`
	Generator string                 `json:"generator,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
	IsPrimary bool                   `json:"is_primary,omitempty"`
}

// Index represents a database index
type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique,omitempty"`
}

// Relation represents a relationship between tables
type Relation struct {
	Type       string `json:"type"`
	FromTable  string `json:"from_table"`
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
	OnDelete   string `json:"on_delete,omitempty"`
	OnUpdate   string `json:"on_update,omitempty"`
}

// Validate validates the schema
func (s *Schema) Validate() error {
	if len(s.Tables) == 0 {
		return fmt.Errorf("schema must contain at least one table")
	}

	// Track table names for relation validation
	tableNames := make(map[string]bool)

	for _, table := range s.Tables {
		if table.Name == "" {
			return fmt.Errorf("table name cannot be empty")
		}
		if table.Count <= 0 {
			return fmt.Errorf("table %s must have a positive count", table.Name)
		}
		if len(table.Columns) == 0 {
			return fmt.Errorf("table %s must have at least one column", table.Name)
		}

		tableNames[table.Name] = true

		// Track column names for index validation
		columnNames := make(map[string]bool)
		for _, col := range table.Columns {
			if col.Name == "" {
				return fmt.Errorf("column name cannot be empty in table %s", table.Name)
			}
			if col.Type == "" {
				return fmt.Errorf("column type cannot be empty for column %s in table %s", col.Name, table.Name)
			}

			columnNames[col.Name] = true

			// Set default industry to "base" if not specified
			if col.Industry == "" {
				col.Industry = "base"
			}

			// Set default generator based on column type if not specified
			if col.Generator == "" {
				switch col.Type {
				case "string", "text", "varchar", "char":
					col.Generator = "word"
				case "integer", "int", "bigint", "smallint":
					col.Generator = "number"
					if col.Params == nil {
						col.Params = map[string]interface{}{
							"min": 0,
							"max": 100,
						}
					}
				case "float", "decimal", "numeric":
					col.Generator = "float"
					if col.Params == nil {
						col.Params = map[string]interface{}{
							"min":       0.0,
							"max":       100.0,
							"precision": 2,
						}
					}
				case "boolean", "bool":
					col.Generator = "boolean"
				case "datetime", "timestamp":
					col.Generator = "datetime"
					if col.Params == nil {
						col.Params = map[string]interface{}{
							"format": "2006-01-02 15:04:05",
						}
					}
				case "date":
					col.Generator = "date"
					if col.Params == nil {
						col.Params = map[string]interface{}{
							"format": "2006-01-02",
						}
					}
				default:
					col.Generator = "word"
				}
			}

			// Validate industry and generator using the global registry
			generatorFunc, err := generators.GlobalRegistry.GetGenerator(col.Industry, col.Generator)
			if err != nil {
				return fmt.Errorf("invalid industry '%s' or generator '%s' for column %s in table %s: %w", col.Industry, col.Generator, col.Name, table.Name, err)
			}

			// Try to generate a test value to ensure the generator works
			_, err = generatorFunc(col.Params)
			if err != nil {
				return fmt.Errorf("failed to generate test value for generator '%s' in industry '%s' for column %s in table %s: %w", col.Generator, col.Industry, col.Name, table.Name, err)
			}
		}

		// Validate indexes
		for _, index := range table.Indexes {
			if index.Name == "" {
				return fmt.Errorf("index name cannot be empty in table %s", table.Name)
			}
			if len(index.Columns) == 0 {
				return fmt.Errorf("index %s must reference at least one column in table %s", index.Name, table.Name)
			}
			for _, colName := range index.Columns {
				if !columnNames[colName] {
					return fmt.Errorf("index %s references non-existent column %s in table %s", index.Name, colName, table.Name)
				}
			}
		}
	}

	// Validate relations
	for _, rel := range s.Relations {
		if rel.Type == "" {
			return fmt.Errorf("relation type cannot be empty")
		}
		if rel.FromTable == "" {
			return fmt.Errorf("from_table cannot be empty in relation")
		}
		if rel.FromColumn == "" {
			return fmt.Errorf("from_column cannot be empty in relation")
		}
		if rel.ToTable == "" {
			return fmt.Errorf("to_table cannot be empty in relation")
		}
		if rel.ToColumn == "" {
			return fmt.Errorf("to_column cannot be empty in relation")
		}

		if !tableNames[rel.FromTable] {
			return fmt.Errorf("relation references non-existent table %s", rel.FromTable)
		}
		if !tableNames[rel.ToTable] {
			return fmt.Errorf("relation references non-existent table %s", rel.ToTable)
		}
	}

	return nil
}

// GetForeignKeyInfo returns the referenced table and column if the given column is a foreign key
func (s *Schema) GetForeignKeyInfo(tableName, columnName string) (string, string, bool) {
	for _, rel := range s.Relations {
		if rel.FromTable == tableName && rel.FromColumn == columnName {
			return rel.ToTable, rel.ToColumn, true
		}
	}
	return "", "", false
}

// GenerateData generates mock data for the schema
func (s *Schema) GenerateData() (map[string][]map[string]interface{}, error) {
	tableData := make(map[string][]map[string]interface{})

	// First, determine the order of tables based on their foreign key dependencies
	tableOrder := make([]string, 0, len(s.Tables))
	processed := make(map[string]bool)

	// Helper function to get tables that this table depends on
	getDependencies := func(tableName string) []string {
		deps := make([]string, 0)
		for _, table := range s.Tables {
			if table.Name == tableName {
				for _, col := range table.Columns {
					if col.Generator == "foreign" {
						if refTable, ok := col.Params["table"].(string); ok {
							deps = append(deps, refTable)
						}
					}
				}
				break
			}
		}
		return deps
	}

	// Helper function to process a table and its dependencies
	var processTable func(tableName string) error
	processTable = func(tableName string) error {
		if processed[tableName] {
			return nil
		}

		// Process dependencies first
		deps := getDependencies(tableName)
		for _, dep := range deps {
			if err := processTable(dep); err != nil {
				return err
			}
		}

		// Add this table to the order
		tableOrder = append(tableOrder, tableName)
		processed[tableName] = true
		return nil
	}

	// Process all tables to determine order
	for _, table := range s.Tables {
		if err := processTable(table.Name); err != nil {
			return nil, err
		}
	}

	// Now generate data in the determined order
	for _, tableName := range tableOrder {
		// Find the table
		var table *Table
		for _, t := range s.Tables {
			if t.Name == tableName {
				table = &t
				break
			}
		}

		if table == nil {
			return nil, fmt.Errorf("table %s not found", tableName)
		}

		// Generate data for this table
		data := make([]map[string]interface{}, table.Count)
		for i := 0; i < table.Count; i++ {
			row := make(map[string]interface{})

			// First generate primary keys and non-foreign key columns
			for _, col := range table.Columns {
				if col.Generator != "foreign" {
					generatorFunc, err := generators.GlobalRegistry.GetGenerator(col.Industry, col.Generator)
					if err != nil {
						return nil, fmt.Errorf("failed to get generator for column %s: %w", col.Name, err)
					}
					value, err := generatorFunc(col.Params)
					if err != nil {
						return nil, fmt.Errorf("failed to generate value for column %s: %w", col.Name, err)
					}
					row[col.Name] = value
				}
			}

			// Then handle foreign keys
			for _, col := range table.Columns {
				if col.Generator == "foreign" {
					refTable := col.Params["table"].(string)
					refCol := col.Params["column"].(string)

					// Verify referenced table exists and has data
					refData, exists := tableData[refTable]
					if !exists {
						return nil, fmt.Errorf("referenced table %s not found", refTable)
					}

					if len(refData) == 0 {
						return nil, fmt.Errorf("no data found in referenced table %s", refTable)
					}

					// Get all values from the referenced column
					refValues := make([]interface{}, 0)
					for _, refRow := range refData {
						if val, ok := refRow[refCol]; ok {
							refValues = append(refValues, val)
						}
					}

					if len(refValues) == 0 {
						return nil, fmt.Errorf("no values found in referenced table %s column %s", refTable, refCol)
					}

					// Use a value from the referenced table
					row[col.Name] = refValues[i%len(refValues)]
				}
			}

			data[i] = row
		}
		tableData[table.Name] = data
	}

	return tableData, nil
}
