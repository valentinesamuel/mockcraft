package generators

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"github.com/valentinesamuel/mockcraft/internal/generators/registry"
)

// GlobalRegistry is the global registry for all generators
var GlobalRegistry = registry.NewIndustryRegistry()

// Generator is the interface that all data generators must implement
type Generator interface {
	// Generate produces a single value based on the generator's type and parameters
	Generate(params map[string]interface{}) (interface{}, error)
}

func init() {
	// Register base generators
	GlobalRegistry.RegisterGenerator("base", "uuid", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.UUID(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "firstname", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.FirstName(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "lastname", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.LastName(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "email", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Email(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "phone", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Phone(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "address", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Address().Address, nil
	})
	GlobalRegistry.RegisterGenerator("base", "company", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Company(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "job_title", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.JobTitle(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "date", func(params map[string]interface{}) (interface{}, error) {
		start := time.Now().AddDate(-1, 0, 0)
		end := time.Now()
		return gofakeit.DateRange(start, end).Format("2006-01-02"), nil
	})
	GlobalRegistry.RegisterGenerator("base", "datetime", func(params map[string]interface{}) (interface{}, error) {
		start := time.Now().AddDate(-1, 0, 0)
		end := time.Now()
		return gofakeit.DateRange(start, end).Format(time.RFC3339), nil
	})
	GlobalRegistry.RegisterGenerator("base", "time", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Date().Format("15:04:05"), nil
	})
	GlobalRegistry.RegisterGenerator("base", "timestamp", func(params map[string]interface{}) (interface{}, error) {
		start := time.Now().AddDate(-1, 0, 0)
		end := time.Now()
		return gofakeit.DateRange(start, end), nil
	})
	GlobalRegistry.RegisterGenerator("base", "number", func(params map[string]interface{}) (interface{}, error) {
		min := 0
		max := 100
		if minVal, ok := params["min"].(int); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(int); ok {
			max = maxVal
		}
		return rand.Intn(max-min+1) + min, nil
	})
	GlobalRegistry.RegisterGenerator("base", "float", func(params map[string]interface{}) (interface{}, error) {
		min := 0.0
		max := 100.0
		if minVal, ok := params["min"].(float64); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(float64); ok {
			max = maxVal
		}
		return min + rand.Float64()*(max-min), nil
	})
	GlobalRegistry.RegisterGenerator("base", "boolean", func(params map[string]interface{}) (interface{}, error) {
		return rand.Float32() < 0.5, nil
	})
	GlobalRegistry.RegisterGenerator("base", "text", func(params map[string]interface{}) (interface{}, error) {
		min := 10
		max := 100
		if minVal, ok := params["min"].(int); ok {
			min = minVal
		}
		if maxVal, ok := params["max"].(int); ok {
			max = maxVal
		}
		length := min + rand.Intn(max-min+1)
		chars := make([]string, length)
		for i := 0; i < length; i++ {
			chars[i] = gofakeit.RandomString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})
		}
		return strings.Join(chars, ""), nil
	})
	GlobalRegistry.RegisterGenerator("base", "paragraph", func(params map[string]interface{}) (interface{}, error) {
		count := 1
		if countVal, ok := params["count"].(int); ok {
			count = countVal
		}
		return gofakeit.Paragraph(count, count, 10, "\n"), nil
	})
	GlobalRegistry.RegisterGenerator("base", "sentence", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Sentence(10), nil
	})
	GlobalRegistry.RegisterGenerator("base", "word", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Word(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "char", func(params map[string]interface{}) (interface{}, error) {
		return string(gofakeit.RandomString([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"})[0]), nil
	})
	GlobalRegistry.RegisterGenerator("base", "url", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.URL(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "ip", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.IPv4Address(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "domain", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.DomainName(), nil
	})
	GlobalRegistry.RegisterGenerator("base", "username", func(params map[string]interface{}) (interface{}, error) {
		return gofakeit.Username(), nil
	})
}

// GenerateRow generates a single row of data for a table, handling relationships
func GenerateRow(table *types.Table, tableIDs map[string][]string, relations []types.Relationship, rowIndex int) (map[string]interface{}, error) {
	row := make(map[string]interface{})

	// Generate values for all columns
	for _, col := range table.Columns {
		// Check if this column is marked as a foreign key
		if col.IsForeign {
			// Handle foreign key relationship
			if col.Params != nil {
				if refTable, ok := col.Params["table"].(string); ok {
					// Get the referenced table's IDs
					referencedIDs := tableIDs[refTable]
					if len(referencedIDs) == 0 {
						log.Printf("Warning: No IDs found for referenced table %s", refTable)
						// Don't skip - this will cause NOT NULL constraint error
						// Instead, we should ensure parent tables are generated first
						return nil, fmt.Errorf("no IDs available for foreign key reference to table %s", refTable)
					}

					// Select a random ID from the referenced table
					randomIndex := rand.Intn(len(referencedIDs))
					row[col.Name] = referencedIDs[randomIndex]
				} else {
					return nil, fmt.Errorf("foreign key column %s missing 'table' parameter", col.Name)
				}
			} else {
				return nil, fmt.Errorf("foreign key column %s missing parameters", col.Name)
			}
		} else {
			// Generate regular column value
			var generatorFunc func(map[string]interface{}) (interface{}, error)
			var err error

			if col.Industry != "" {
				generatorFunc, err = GlobalRegistry.GetGenerator(col.Industry, col.Generator)
				if err != nil {
					log.Printf("Warning: Failed to get generator for industry '%s', generator '%s': %v", col.Industry, col.Generator, err)
					// Fall back to base generator
					generatorFunc, _ = GlobalRegistry.GetGenerator("base", GetDefaultGeneratorNameByColumnType(col))
				}
			} else {
				generatorFunc, _ = GlobalRegistry.GetGenerator("base", GetDefaultGeneratorNameByColumnType(col))
			}

			if generatorFunc != nil {
				value, err := generatorFunc(col.Params)
				if err != nil {
					log.Printf("Warning: Failed to generate value for column %s.%s with generator '%s': %v", table.Name, col.Name, col.Generator, err)
					return nil, fmt.Errorf("failed to generate value for column %s: %v", col.Name, err)
				}
				row[col.Name] = value
			} else {
				return nil, fmt.Errorf("no generator found for column %s", col.Name)
			}
		}
	}

	return row, nil
}

// GetTableGenerationOrder returns tables in dependency order (parents first)
func GetTableGenerationOrder(tables []types.Table) []types.Table {
	// Create a map for quick lookup
	tableMap := make(map[string]*types.Table)
	for i := range tables {
		tableMap[tables[i].Name] = &tables[i]
	}

	// Track dependencies
	dependencies := make(map[string][]string)
	for _, table := range tables {
		dependencies[table.Name] = []string{}
		for _, col := range table.Columns {
			if col.IsForeign && col.Params != nil {
				if refTable, ok := col.Params["table"].(string); ok {
					dependencies[table.Name] = append(dependencies[table.Name], refTable)
				}
			}
		}
	}

	// Topological sort
	var result []types.Table
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(string) bool
	visit = func(tableName string) bool {
		if visiting[tableName] {
			log.Printf("Warning: Circular dependency detected involving table %s", tableName)
			return false
		}
		if visited[tableName] {
			return true
		}

		visiting[tableName] = true
		for _, dep := range dependencies[tableName] {
			if !visit(dep) {
				return false
			}
		}
		visiting[tableName] = false
		visited[tableName] = true

		if table, exists := tableMap[tableName]; exists {
			result = append(result, *table)
		}
		return true
	}

	for _, table := range tables {
		visit(table.Name)
	}

	return result
}

func GetDefaultGeneratorNameByColumnType(col types.Column) string {
	switch col.Type {
	case "string", "text":
		return "text"
	case "uuid":
		return "uuid"
	case "integer":
		return "number"
	case "decimal", "float":
		return "float" // Changed from "decimal" to "float"
	case "boolean":
		return "boolean"
	case "timestamp", "datetime", "date":
		return "timestamp"
	default:
		log.Printf("Warning: No fallback generator for type '%s' for column %s.", col.Type, col.Name)
		return "text" // Changed from "nil" to "text" as fallback
	}
}
