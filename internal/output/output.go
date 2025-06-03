package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
				record = append(record, fmt.Sprintf("%v", row[k]))
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
func WriteSQL(dir, tableName string, data []map[string]interface{}) error {
	// Create file
	filename := filepath.Join(dir, fmt.Sprintf("%s.sql", tableName))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write INSERT statements
	for _, row := range data {
		columns := make([]string, 0, len(row))
		values := make([]string, 0, len(row))
		for col, val := range row {
			columns = append(columns, fmt.Sprintf("`%s`", col))
			values = append(values, fmt.Sprintf("%v", val))
		}
		fmt.Fprintf(file, "INSERT INTO `%s` (%s) VALUES (%s);\n",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(values, ", "),
		)
	}

	return nil
}
