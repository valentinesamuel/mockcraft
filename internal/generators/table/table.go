package table

import (
	"fmt"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"github.com/valentinesamuel/mockcraft/internal/generators"
)

// GenerateTableData generates data for a table
func GenerateTableData(table types.Table) ([]map[string]interface{}, error) {
	// Create data slice
	data := make([]map[string]interface{}, len(table.Data))
	for i, row := range table.Data {
		// Copy existing data
		data[i] = make(map[string]interface{})
		for k, v := range row {
			data[i][k] = v
		}

		// Generate missing values
		for _, col := range table.Columns {
			if _, ok := data[i][col.Name]; !ok {
				// Get generator
				generator, err := generators.Get(col.Type)
				if err != nil {
					return nil, fmt.Errorf("failed to get generator for column %s: %w", col.Name, err)
				}

				// Generate value
				value, err := generator.Generate(nil)
				if err != nil {
					return nil, fmt.Errorf("failed to generate value for column %s: %w", col.Name, err)
				}

				data[i][col.Name] = value
			}
		}
	}

	return data, nil
}
