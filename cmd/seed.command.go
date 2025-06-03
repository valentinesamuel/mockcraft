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
	"time"

	"github.com/spf13/cobra"
	"github.com/valentinesamuel/mockcraft/internal/config"
	"github.com/valentinesamuel/mockcraft/internal/database"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
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

func seedDatabase(schema *types.Schema) error {
	// Parse database DSN
	dbConfig, err := parseDatabaseURL(seedDB)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Create database connection
	db, err := database.NewDatabase(dbConfig)
	if err != nil {
		return fmt.Errorf("failed to create database connection: %w", err)
	}
	defer db.Close()

	// Connect to database
	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Process each table
	for _, table := range schema.Tables {
		if seedDryRun {
			log.Printf("Would seed table %s with %d rows", table.Name, len(table.Data))
			continue
		}

		// Create table
		if err := db.CreateTable(ctx, table.Name, table.Columns); err != nil {
			return fmt.Errorf("failed to create table %s: %v", table.Name, err)
		}

		// Create indexes
		for _, index := range table.Indexes {
			if err := db.CreateIndex(ctx, table.Name, index); err != nil {
				return fmt.Errorf("failed to create index %s on table %s: %v", index.Name, table.Name, err)
			}
		}

		// Insert data
		if len(table.Data) > 0 {
			if err := db.InsertData(ctx, table.Name, table.Data); err != nil {
				return fmt.Errorf("failed to insert data into table %s: %v", table.Name, err)
			}
		}

		log.Printf("Seeded table %s with %d rows", table.Name, len(table.Data))
	}

	return nil
}

func generateFiles(schema *types.Schema) error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Process each table
	for _, table := range schema.Tables {
		if seedDryRun {
			log.Printf("Would generate %s file for table %s with %d rows", seedOutput, table.Name, len(table.Data))
			continue
		}

		// Generate file based on output format
		filename := filepath.Join(seedDir, fmt.Sprintf("%s.%s", table.Name, seedOutput))
		switch strings.ToLower(seedOutput) {
		case "csv":
			if err := writeCSV(filename, table.Columns, table.Data); err != nil {
				return fmt.Errorf("failed to write CSV file: %v", err)
			}
		case "json":
			if err := writeJSON(filename, table.Columns, table.Data); err != nil {
				return fmt.Errorf("failed to write JSON file: %v", err)
			}
		case "sql":
			if err := writeSQL(filename, table); err != nil {
				return fmt.Errorf("failed to write SQL file: %v", err)
			}
		default:
			return fmt.Errorf("unsupported output format: %s", seedOutput)
		}

		log.Printf("Generated %s file for table %s with %d rows", seedOutput, table.Name, len(table.Data))
	}

	return nil
}

func writeCSV(filename string, columns []types.Column, data []map[string]interface{}) error {
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
		record := make([]string, len(columns))
		for i, col := range columns {
			record[i] = fmt.Sprintf("%v", row[col.Name])
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

func writeJSON(filename string, columns []types.Column, data []map[string]interface{}) error {
	// Write JSON file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func writeSQL(filename string, table types.Table) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write CREATE TABLE statement
	fmt.Fprintf(file, "CREATE TABLE IF NOT EXISTS `%s` (\n", table.Name)
	for i, col := range table.Columns {
		fmt.Fprintf(file, "  `%s` %s", col.Name, col.Type)
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

	// Write INSERT statements
	for _, row := range table.Data {
		columns := make([]string, 0, len(row))
		values := make([]string, 0, len(row))
		for col, val := range row {
			columns = append(columns, fmt.Sprintf("`%s`", col))
			values = append(values, fmt.Sprintf("%v", val))
		}
		fmt.Fprintf(file, "INSERT INTO `%s` (%s) VALUES (%s);\n",
			table.Name,
			strings.Join(columns, ", "),
			strings.Join(values, ", "),
		)
	}

	return nil
}

// parseDatabaseURL parses a database URL into a database configuration
func parseDatabaseURL(url string) (types.Config, error) {
	// TODO: Implement URL parsing for different database drivers
	// For now, return a basic configuration
	return types.Config{
		Driver:          "postgres", // Default to PostgreSQL
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: time.Minute * 5,
	}, nil
}
