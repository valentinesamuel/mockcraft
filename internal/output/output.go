package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

// WriteCSV writes data to a CSV file
func WriteCSV(dir, tableName string, data []map[string]interface{}) error {
	// Create file
	filename := filepath.Join(dir, fmt.Sprintf("%s.csv", tableName))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Create writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if len(data) > 0 {
		header := make([]string, 0, len(data[0]))
		for k := range data[0] {
			header = append(header, k)
		}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}

		// Write data
		for _, row := range data {
			record := make([]string, 0, len(row))
			for _, k := range header {
				record = append(record, formatValue(row[k]))
			}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write record: %w", err)
			}
		}
	}

	return nil
}

// WriteJSON writes data to a JSON file
func WriteJSON(dir, tableName string, data []map[string]interface{}) error {
	// Create file
	filename := filepath.Join(dir, fmt.Sprintf("%s.json", tableName))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write JSON
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	return nil
}

// WriteSQL writes data to a SQL file
func WriteSQL(filename string, table types.Table) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write CREATE TABLE statement
	fmt.Fprintf(file, "CREATE TABLE IF NOT EXISTS %s (\n", table.Name)
	for i, col := range table.Columns {
		// Convert type to proper SQL type
		sqlType := getSQLType(col.Type)
		fmt.Fprintf(file, "  %s %s", col.Name, sqlType)
		if col.IsPrimary {
			fmt.Fprintf(file, " PRIMARY KEY")
		}
		if !col.IsNullable {
			fmt.Fprintf(file, " NOT NULL")
		}
		if col.IsUnique {
			fmt.Fprintf(file, " UNIQUE")
		}
		if col.Default != nil {
			fmt.Fprintf(file, " DEFAULT %v", col.Default)
		}
		if i < len(table.Columns)-1 {
			fmt.Fprintf(file, ",")
		}
		fmt.Fprintf(file, "\n")
	}
	fmt.Fprintf(file, ");\n\n")

	// Get column names from first row and sort them for consistent order
	var columns []string
	if len(table.Data) > 0 {
		for col := range table.Data[0] {
			columns = append(columns, col)
		}
		sort.Strings(columns)
	}

	// Write INSERT statements
	for _, row := range table.Data {
		values := make([]string, len(columns))
		for i, col := range columns {
			val := row[col]
			values[i] = formatSQLValue(val)
		}
		fmt.Fprintf(file, "INSERT INTO %s (%s) VALUES (%s);\n",
			table.Name,
			strings.Join(columns, ", "),
			strings.Join(values, ", "),
		)
	}

	return nil
}

// getSQLType converts a generic type to SQL type
func getSQLType(typ string) string {
	switch strings.ToLower(typ) {
	case "uuid":
		return "CHAR(36)"
	case "string":
		return "VARCHAR(255)"
	case "text":
		return "TEXT"
	case "integer":
		return "INT"
	case "decimal":
		return "DECIMAL(10,2)"
	case "boolean":
		return "BOOLEAN"
	case "timestamp":
		return "TIMESTAMP"
	default:
		return typ
	}
}

// formatValue formats a value for CSV output
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int32, int64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%f", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case time.Time:
		return val.Format(time.RFC3339)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatSQLValue formats a value for SQL output
func formatSQLValue(val interface{}) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case string:
		// Escape single quotes and wrap in single quotes
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	case bool:
		if v {
			return "TRUE"
		}
		return "FALSE"
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", v)
	default:
		// For any other type, convert to string and quote
		return "'" + fmt.Sprintf("%v", v) + "'"
	}
}
