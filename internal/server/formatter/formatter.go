package formatter

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/valentinesamuel/mockcraft/internal/server/schema"
)

// Formatter defines the interface for data formatters
type Formatter interface {
	Format(schema *schema.Schema, data map[string][][]interface{}, format string) (string, error)
}

// CSVFormatter implements the Formatter interface for CSV output
type CSVFormatter struct {
	outputDir string
}

// NewCSVFormatter creates a new CSVFormatter instance
func NewCSVFormatter() *CSVFormatter {
	return &CSVFormatter{}
}

// Format generates CSV files based on the schema and data
func (f *CSVFormatter) Format(schema *schema.Schema, data map[string][][]interface{}, format string) (string, error) {
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
