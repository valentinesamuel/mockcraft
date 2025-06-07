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

	for _, table := range schema.Tables {
		filePath := filepath.Join(outputDir, table.Name+".csv")
		file, err := os.Create(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to create CSV file: %w", err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write header
		header := make([]string, len(table.Columns))
		for i, col := range table.Columns {
			header[i] = col.Name
		}
		if err := writer.Write(header); err != nil {
			return "", fmt.Errorf("failed to write CSV header: %w", err)
		}

		// Write data
		rows := data[table.Name]
		for _, row := range rows {
			record := make([]string, len(row))
			for i, val := range row {
				record[i] = fmt.Sprint(val)
			}
			if err := writer.Write(record); err != nil {
				return "", fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	return outputDir, nil
}

// formatJSON generates JSON files
func (f *Formatter) formatJSON(schema *schema.Schema, data map[string][][]interface{}) (string, error) {
	outputDir := filepath.Join(f.outputDir, "json")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, table := range schema.Tables {
		filePath := filepath.Join(outputDir, table.Name+".json")
		file, err := os.Create(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to create JSON file: %w", err)
		}
		defer file.Close()

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
			return "", fmt.Errorf("failed to write JSON data: %w", err)
		}
	}

	return outputDir, nil
}

// formatSQL generates SQL files
func (f *Formatter) formatSQL(schema *schema.Schema, data map[string][][]interface{}) (string, error) {
	outputDir := filepath.Join(f.outputDir, "sql")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, table := range schema.Tables {
		filePath := filepath.Join(outputDir, table.Name+".sql")
		file, err := os.Create(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to create SQL file: %w", err)
		}
		defer file.Close()

		// Write CREATE TABLE statement
		createTable := fmt.Sprintf("CREATE TABLE %s (\n", table.Name)
		columns := make([]string, len(table.Columns))
		for i, col := range table.Columns {
			nullable := "NOT NULL"
			if col.Nullable {
				nullable = "NULL"
			}
			unique := ""
			if col.Unique {
				unique = "UNIQUE"
			}
			columns[i] = fmt.Sprintf("  %s %s %s %s", col.Name, col.Type, nullable, unique)
		}
		createTable += strings.Join(columns, ",\n") + "\n);\n\n"
		if _, err := file.WriteString(createTable); err != nil {
			return "", fmt.Errorf("failed to write CREATE TABLE: %w", err)
		}

		// Write INSERT statements
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
				return "", fmt.Errorf("failed to write INSERT: %w", err)
			}
		}
	}

	return outputDir, nil
}
