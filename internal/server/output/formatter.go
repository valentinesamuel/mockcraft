package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/valentinesamuel/mockcraft/internal/server/schema"
)

// Formatter handles output file generation
type Formatter struct {
	outputDir string
}

// New creates a new formatter
func New(outputDir string) *Formatter {
	return &Formatter{
		outputDir: outputDir,
	}
}

// Format generates output files based on the schema and data
func (f *Formatter) Format(schema *schema.Schema, data map[string][][]interface{}, format string) (string, error) {
	switch strings.ToLower(format) {
	case "csv":
		return f.formatCSV(schema, data)
	case "json":
		return f.formatJSON(schema, data)
	case "sql":
		return f.formatSQL(schema, data)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

// formatCSV generates CSV files
func (f *Formatter) formatCSV(schema *schema.Schema, data map[string][][]interface{}) (string, error) {
	outputDir := filepath.Join(f.outputDir, "csv")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Process each table
	for _, table := range schema.Tables {
		filePath := filepath.Join(outputDir, table.Name+".csv")
		file, err := os.Create(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to create CSV file for table %s: %w", table.Name, err)
		}

		writer := csv.NewWriter(file)

		// Write header
		header := make([]string, len(table.Columns))
		for i, col := range table.Columns {
			header[i] = col.Name
		}
		if err := writer.Write(header); err != nil {
			file.Close()
			return "", fmt.Errorf("failed to write CSV header for table %s: %w", table.Name, err)
		}

		// Write data
		rows := data[table.Name]
		for _, row := range rows {
			record := make([]string, len(row))
			for i, val := range row {
				record[i] = fmt.Sprint(val)
			}
			if err := writer.Write(record); err != nil {
				file.Close()
				return "", fmt.Errorf("failed to write CSV row for table %s: %w", table.Name, err)
			}
		}

		writer.Flush()
		file.Close()
	}

	return outputDir, nil
}

// formatJSON generates JSON files
func (f *Formatter) formatJSON(schema *schema.Schema, data map[string][][]interface{}) (string, error) {
	outputDir := filepath.Join(f.outputDir, "json")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Process each table
	for _, table := range schema.Tables {
		filePath := filepath.Join(outputDir, table.Name+".json")
		file, err := os.Create(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to create JSON file for table %s: %w", table.Name, err)
		}

		// Convert data to map format
		records := make([]map[string]interface{}, len(data[table.Name]))
		for i, row := range data[table.Name] {
			record := make(map[string]interface{})
			for j, col := range table.Columns {
				record[col.Name] = row[j]
			}
			records[i] = record
		}

		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(records); err != nil {
			file.Close()
			return "", fmt.Errorf("failed to write JSON data for table %s: %w", table.Name, err)
		}

		file.Close()
	}

	return outputDir, nil
}

// formatSQL generates SQL files
func (f *Formatter) formatSQL(schema *schema.Schema, data map[string][][]interface{}) (string, error) {
	outputDir := filepath.Join(f.outputDir, "sql")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create a single SQL file for all tables
	filePath := filepath.Join(outputDir, "schema.sql")
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create SQL file: %w", err)
	}
	defer file.Close()

	// Write CREATE TABLE statements
	for _, table := range schema.Tables {
		createTable := fmt.Sprintf("CREATE TABLE %s (\n", table.Name)
		columns := make([]string, len(table.Columns))
		for i, col := range table.Columns {
			columnDef := fmt.Sprintf("  %s %s", col.Name, col.Type)
			if col.IsPrimary {
				columnDef += " PRIMARY KEY"
			}
			if !col.IsNullable {
				columnDef += " NOT NULL"
			}
			columns[i] = columnDef
		}
		createTable += strings.Join(columns, ",\n") + "\n);\n\n"
		if _, err := file.WriteString(createTable); err != nil {
			return "", fmt.Errorf("failed to write CREATE TABLE for %s: %w", table.Name, err)
		}
	}

	// Write INSERT statements
	for _, table := range schema.Tables {
		rows := data[table.Name]
		for _, row := range rows {
			values := make([]string, len(row))
			for i, val := range row {
				switch v := val.(type) {
				case string:
					values[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
				case nil:
					values[i] = "NULL"
				default:
					values[i] = fmt.Sprint(v)
				}
			}
			insert := fmt.Sprintf("INSERT INTO %s VALUES (%s);\n",
				table.Name, strings.Join(values, ", "))
			if _, err := file.WriteString(insert); err != nil {
				return "", fmt.Errorf("failed to write INSERT for %s: %w", table.Name, err)
			}
		}
	}

	return outputDir, nil
}
