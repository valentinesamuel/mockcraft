package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/valentinesamuel/mockcraft/internal/config"
	"github.com/valentinesamuel/mockcraft/internal/database"
	"github.com/valentinesamuel/mockcraft/internal/database/postgres"
	"github.com/valentinesamuel/mockcraft/internal/generators"
)

var (
	seedConfigPath string
	seedDB         string
	seedOutput     string
	seedDir        string
	seedCount      int
	seedDryRun     bool
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed a database or generate data files from a schema",
	Long: `Seed a database with fake data based on a YAML schema configuration.
Example:
mockcraft seed --config schema.yaml --db postgres://...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load schema
		schema, err := config.LoadSchema(seedConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load schema: %v", err)
		}

		// Override count if specified
		if seedCount > 0 {
			for i := range schema.Tables {
				schema.Tables[i].Count = seedCount
			}
		}

		// Handle database seeding
		if seedDB != "" {
			if err := seedDatabase(schema); err != nil {
				return fmt.Errorf("failed to seed database: %v", err)
			}
		}

		// Handle file output
		if seedOutput != "" {
			if err := generateFiles(schema); err != nil {
				return fmt.Errorf("failed to generate files: %v", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)

	seedCmd.Flags().StringVarP(&seedConfigPath, "config", "c", "", "Path to schema YAML file")
	seedCmd.Flags().StringVar(&seedDB, "db", "", "Database DSN (e.g. postgres://user:pass@host:port/db)")
	seedCmd.Flags().StringVarP(&seedOutput, "output", "o", "", "Output format (csv, json, sql)")
	seedCmd.Flags().StringVar(&seedDir, "dir", "", "Output directory for generated files")
	seedCmd.Flags().IntVar(&seedCount, "count", 0, "Override row count for all tables")
	seedCmd.Flags().BoolVar(&seedDryRun, "dry-run", false, "Print actions without inserting or writing files")
}

func seedDatabase(schema *config.Schema) error {
	// Create seeder based on database type
	var seeder database.Seeder
	var err error

	switch {
	case strings.HasPrefix(seedDB, "postgres://"):
		seeder, err = postgres.NewSeeder(seedDB)
	default:
		return fmt.Errorf("unsupported database type")
	}

	if err != nil {
		return fmt.Errorf("failed to create seeder: %v", err)
	}
	defer seeder.Close()

	// Connect to database
	ctx := context.Background()
	if err := seeder.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Process each table
	for _, table := range schema.Tables {
		if seedDryRun {
			log.Printf("Would seed table %s with %d rows", table.Name, table.Count)
			continue
		}

		// Generate data for table
		data := make([][]interface{}, table.Count)
		for i := 0; i < table.Count; i++ {
			row := make([]interface{}, len(table.Columns))
			for j, col := range table.Columns {
				generator, err := generators.Get(col.Generator)
				if err != nil {
					return fmt.Errorf("failed to get generator '%s' for column '%s': %v", col.Generator, col.Name, err)
				}

				value, err := generator.Generate(col.Params)
				if err != nil {
					return fmt.Errorf("failed to generate value for column '%s': %v", col.Name, err)
				}

				row[j] = value
			}
			data[i] = row
		}

		if err := seeder.SeedTable(ctx, &table, data); err != nil {
			return fmt.Errorf("failed to seed table %s: %v", table.Name, err)
		}

		log.Printf("Seeded table %s with %d rows", table.Name, table.Count)
	}

	return nil
}

func generateFiles(schema *config.Schema) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Process each table
	for _, table := range schema.Tables {
		if seedDryRun {
			log.Printf("Would generate %s file for table %s with %d rows", seedOutput, table.Name, table.Count)
			continue
		}

		// Generate data for table
		data := make([][]interface{}, table.Count)
		for i := 0; i < table.Count; i++ {
			row := make([]interface{}, len(table.Columns))
			for j, col := range table.Columns {
				generator, err := generators.Get(col.Generator)
				if err != nil {
					return fmt.Errorf("failed to get generator '%s' for column '%s': %v", col.Generator, col.Name, err)
				}

				value, err := generator.Generate(col.Params)
				if err != nil {
					return fmt.Errorf("failed to generate value for column '%s': %v", col.Name, err)
				}

				row[j] = value
			}
			data[i] = row
		}

		// Generate file based on output format
		filename := filepath.Join(seedDir, fmt.Sprintf("%s.%s", table.Name, seedOutput))
		switch strings.ToLower(seedOutput) {
		case "csv":
			if err := writeCSV(filename, table.Columns, data); err != nil {
				return fmt.Errorf("failed to write CSV file: %v", err)
			}
		case "json":
			if err := writeJSON(filename, table.Columns, data); err != nil {
				return fmt.Errorf("failed to write JSON file: %v", err)
			}
		case "sql":
			if err := writeSQL(filename, table, data); err != nil {
				return fmt.Errorf("failed to write SQL file: %v", err)
			}
		default:
			return fmt.Errorf("unsupported output format: %s", seedOutput)
		}

		log.Printf("Generated %s file for table %s with %d rows", seedOutput, table.Name, table.Count)
	}

	return nil
}

func writeCSV(filename string, columns []config.Column, data [][]interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := make([]string, len(columns))
	for i, col := range columns {
		header[i] = col.Name
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data
	for _, row := range data {
		record := make([]string, len(row))
		for i, val := range row {
			record[i] = fmt.Sprintf("%v", val)
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func writeJSON(filename string, columns []config.Column, data [][]interface{}) error {
	// Convert data to map format
	records := make([]map[string]interface{}, len(data))
	for i, row := range data {
		record := make(map[string]interface{})
		for j, col := range columns {
			record[col.Name] = row[j]
		}
		records[i] = record
	}

	// Write JSON file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(records)
}

func writeSQL(filename string, table config.Table, data [][]interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write table creation SQL
	columnDefs := make([]string, len(table.Columns))
	for i, col := range table.Columns {
		def := fmt.Sprintf("%s %s", col.Name, col.Type)
		if col.PrimaryKey {
			def += " PRIMARY KEY"
		}
		columnDefs[i] = def
	}

	fmt.Fprintf(file, "CREATE TABLE IF NOT EXISTS %s (\n  %s\n);\n\n", table.Name, strings.Join(columnDefs, ",\n  "))

	// Write INSERT statements
	for _, row := range data {
		values := make([]string, len(row))
		for i, val := range row {
			switch v := val.(type) {
			case string:
				values[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
			case nil:
				values[i] = "NULL"
			default:
				values[i] = fmt.Sprintf("%v", v)
			}
		}

		columns := make([]string, len(table.Columns))
		for i, col := range table.Columns {
			columns[i] = col.Name
		}

		fmt.Fprintf(file, "INSERT INTO %s (%s) VALUES (%s);\n",
			table.Name,
			strings.Join(columns, ", "),
			strings.Join(values, ", "))
	}

	return nil
}
