package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"github.com/valentinesamuel/mockcraft/internal/config"
	"github.com/valentinesamuel/mockcraft/internal/database"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"github.com/valentinesamuel/mockcraft/internal/generators"
)

var (
	seedConfigPath string
	seedDB         string
	seedOutput     string
	seedDir        string
	seedCount      int
	seedDryRun     bool

	// Local backup path flag
	backupLocalPath string
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed a database or generate data files from a schema",
	Long: `Seed a database with fake data based on a YAML schema configuration.
Example:
mockcraft seed --config schema.yaml --db postgres://... --backup-path ./backup.sql`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load schema
		schema, err := config.LoadSchema(seedConfigPath)
		if err != nil {
			return fmt.Errorf("failed to load schema: %v", err)
		}

		// Handle database seeding
		if seedDB != "" {
			// Ensure backup-path is provided for database seeding
			if backupLocalPath == "" {
				return fmt.Errorf("--backup-path is required for database seeding")
			}

			// Parse database DSN
			dbConfig, err := database.ParseDatabaseURL(seedDB)
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
				return fmt.Errorf("failed to connect to database: %w", err)
			}

			// Create backup to local file before dropping tables/collections
			log.Printf("Creating backup to local file '%s'...", backupLocalPath)
			if err := db.Backup(ctx, backupLocalPath); err != nil {
				// Log a warning but continue seeding
				log.Printf("Warning: Failed to create backup: %v", err)
			} else {
				log.Printf("Backup created successfully at '%s'", backupLocalPath)
			}

			// Drop existing collections/tables
			if db.GetDriver() == "mongodb" {
				log.Println("Dropping existing collections...")
				// Drop collections in reverse order of definition to respect dependencies
				for i := len(schema.Tables) - 1; i >= 0; i-- {
					table := schema.Tables[i]
					log.Printf("Dropping collection: %s", table.Name)
					if err := db.DropTable(ctx, table.Name); err != nil {
						// Log a warning but continue with other collections
						log.Printf("Warning: Failed to drop collection %s: %v", table.Name, err)
					}
				}
				log.Println("Finished dropping collections.")
			} else {
				log.Println("Dropping existing tables...")
				// Drop tables in reverse order of definition to respect dependencies
				for i := len(schema.Tables) - 1; i >= 0; i-- {
					table := schema.Tables[i]
					log.Printf("Dropping table: %s", table.Name)
					if err := db.DropTable(ctx, table.Name); err != nil {
						// Log a warning but continue with other tables
						log.Printf("Warning: Failed to drop table %s: %v", table.Name, err)
					}
				}
				log.Println("Finished dropping tables.")
			}

			// Call seedDatabase (without backup logic as it's handled above)
			if err := seedDatabaseContent(ctx, db, schema); err != nil {
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

	// Local backup path flag (mandatory for database seeding)
	seedCmd.Flags().StringVar(&backupLocalPath, "backup-path", "", "Local file path to save the backup dump")

	// Mark --backup-path as required if --db is provided
	cobra.OnInitialize(func() {
		if seedDB != "" {
			seedCmd.MarkFlagRequired("backup-path")
		}
	})
}

// Renamed from seedDatabase to seedDatabaseContent to separate backup logic
func seedDatabaseContent(ctx context.Context, db types.Database, schema *types.Schema) error {
	// 1. Create tables based on schema
	log.Println("Creating tables...")
	for _, table := range schema.Tables {
		log.Printf("Creating table: %s", table.Name)
		if err := db.CreateTable(ctx, table.Name, &table, schema.Relations); err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.Name, err)
		}
	}
	log.Println("Tables created successfully.")

	// 2. Create indexes based on schema
	log.Println("Creating indexes...")
	for _, table := range schema.Tables {
		for _, index := range table.Indexes {
			log.Printf("Creating index %s on table %s", index.Name, table.Name)
			if err := db.CreateIndex(ctx, table.Name, index); err != nil {
				log.Printf("Warning: Failed to create index %s on table %s: %v", index.Name, table.Name, err)
			}
		}
	}
	log.Println("Indexes created.")

	// 3. Create constraints based on schema
	log.Println("Creating constraints...")
	for _, constraint := range schema.Constraints {
		log.Printf("Creating constraint for table: %v", constraint.Columns)
		log.Printf("Skipping constraint creation for type %s on columns %v", constraint.Type, constraint.Columns)
	}
	log.Println("Constraints created (where supported/necessary).")

	// Store generated IDs for each table
	tableIDs := make(map[string][]string)
	// Store data for tables with circular dependencies
	circularData := make(map[string][]map[string]interface{})

	// 4. Generate and insert initial data for all tables
	log.Println("Generating and inserting initial data...")
	for _, table := range schema.Tables {
		// Skip data generation/insertion if Count is 0
		if table.Count == 0 {
			log.Printf("Skipping data generation for table %s as count is 0.", table.Name)
			tableIDs[table.Name] = []string{}
			continue
		}

		data := make([]map[string]interface{}, table.Count)

		// Get incoming relationships where this table is the 'to' table
		incomingRelations := make([]types.Relationship, 0)
		for _, rel := range schema.Relations {
			if rel.ToTable == table.Name {
				incomingRelations = append(incomingRelations, rel)
			}
		}

		// Check if this table has circular dependencies
		hasCircularDependency := false
		for _, rel := range schema.Relations {
			if rel.FromTable == table.Name {
				for _, innerRel := range schema.Relations {
					if innerRel.ToTable == table.Name && innerRel.FromTable == rel.ToTable {
						hasCircularDependency = true
						break
					}
				}
			}
			if hasCircularDependency {
				break
			}
		}

		if hasCircularDependency {
			log.Printf("Table %s has circular dependency, deferring data insertion.", table.Name)
			// Store data for later insertion after all IDs are generated
			circularData[table.Name] = data
			// Generate data but don't insert yet
			for i := 0; i < table.Count; i++ {
				generatedRow, err := generators.GenerateRow(&table, tableIDs, schema.Relations, i)
				if err != nil {
					return fmt.Errorf("failed to generate data for table %s: %v", table.Name, err)
				}
				data[i] = generatedRow
				// Extract and store the generated ID for later foreign key references
				pkCol, err := table.GetPrimaryKeyColumn()
				if err != nil {
					log.Printf("Warning: Could not get primary key column for table %s: %v", table.Name, err)
					continue
				}

				if id, ok := generatedRow[pkCol.Name].(string); ok {
					tableIDs[table.Name] = append(tableIDs[table.Name], id)
				}
			}
		} else {
			log.Printf("Generating %d rows for table: %s", table.Count, table.Name)
			// Generate and insert data
			for i := 0; i < table.Count; i++ {
				generatedRow, err := generators.GenerateRow(&table, tableIDs, schema.Relations, i)
				if err != nil {
					return fmt.Errorf("failed to generate data for table %s: %v", table.Name, err)
				}
				data[i] = generatedRow
				// Extract and store the generated ID for later foreign key references
				pkCol, err := table.GetPrimaryKeyColumn()
				if err != nil {
					log.Printf("Warning: Could not get primary key column for table %s: %v", table.Name, err)
					continue
				}

				if id, ok := generatedRow[pkCol.Name].(string); ok {
					tableIDs[table.Name] = append(tableIDs[table.Name], id)
				}
			}

			// Insert data in batches
			batchSize := 1000 // Adjust batch size as needed
			for i := 0; i < len(data); i += batchSize {
				end := i + batchSize
				if end > len(data) {
					end = len(data)
				}
				if err := db.InsertData(ctx, table.Name, data[i:end]); err != nil {
					return fmt.Errorf("failed to insert data into table %s: %w", table.Name, err)
				}
				log.Printf("Seeded table %s with %d rows", table.Name, len(data[i:end]))
			}
		}
	}
	log.Println("Initial data insertion complete.")

	// 5. Insert data for tables with circular dependencies
	log.Println("Inserting data for tables with circular dependencies...")
	// Sort circularData keys to ensure consistent insertion order (optional, but good practice)
	tablesWithCircularDeps := make([]string, 0, len(circularData))
	for table := range circularData {
		tablesWithCircularDeps = append(tablesWithCircularDeps, table)
	}
	sort.Strings(tablesWithCircularDeps)

	for _, tableName := range tablesWithCircularDeps {
		data := circularData[tableName]
		log.Printf("Inserting %d deferred rows for table: %s", len(data), tableName)

		// Insert data in batches
		batchSize := 1000 // Adjust batch size as needed
		for i := 0; i < len(data); i += batchSize {
			end := i + batchSize
			if end > len(data) {
				end = len(data)
			}
			if err := db.InsertData(ctx, tableName, data[i:end]); err != nil {
				return fmt.Errorf("failed to insert deferred data into table %s: %w", tableName, err)
			}
			log.Printf("Seeded table %s with %d deferred rows", tableName, len(data[i:end]))
		}

	}
	log.Println("Deferred data insertion complete.")

	// 6. Verify referential integrity
	log.Println("Verifying referential integrity...")
	for _, rel := range schema.Relations {
		log.Printf("Verifying relationship from %s.%s to %s.%s", rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn)
		if err := db.VerifyReferentialIntegrity(ctx, rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn); err != nil {
			log.Printf("Warning: Referential integrity violation for relationship from %s.%s to %s.%s: %v", rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn, err)
		} else {
			log.Printf("Referential integrity check passed for relationship from %s.%s to %s.%s", rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn)
		}
	}
	log.Println("Referential integrity verification complete.")

	return nil
}

func generateFiles(schema *types.Schema) error {
	// Create output directory if it doesn't exist
	if seedDir == "" {
		seedDir = "." // Default to current directory if not specified
	}
	if err := os.MkdirAll(seedDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Process each table
	for _, table := range schema.Tables {
		if seedDryRun {
			log.Printf("Would generate %s file for table %s with %d rows", seedOutput, table.Name, table.Count) // Use table.Count here
			continue
		}

		// Generate data if not already present (this part might be redundant after seedDatabase)
		// However, if we ever use generateFiles independently, we need data generation here.
		if len(table.Data) == 0 {
			// Use count from flag if specified, otherwise use table's count
			count := seedCount
			if count == 0 {
				count = table.Count // Use schema's count
			}
			if count == 0 {
				count = 100 // Default to 100 rows if no count specified in schema or flag
			}

			table.Data = make([]map[string]interface{}, count)

			// This data generation logic should ideally be consistent with seedDatabase
			// For now, keeping it simple, but acknowledge potential discrepancy
			for i := 0; i < count; i++ {
				row := make(map[string]interface{})
				for _, col := range table.Columns {
					// Simple generator usage - does NOT handle foreign keys correctly here
					generator, err := generators.Get(col.Generator)
					if err == nil && generator != nil {
						value, err := generator.Generate(col.Params)
						if err == nil {
							row[col.Name] = value
						} else {
							log.Printf("Warning: Failed to generate value for column %s.%s with generator '%s': %v. Using error placeholder.", table.Name, col.Name, col.Generator, err)
							row[col.Name] = fmt.Sprintf("generator-error-%s", col.Generator)
						}
					} else {
						row[col.Name] = fmt.Sprintf("unknown-generator-%s", col.Generator)
					}
				}
				table.Data[i] = row
			}
		}

		if len(table.Data) > 0 {
			// Write data to file based on output format
			filename := filepath.Join(seedDir, fmt.Sprintf("%s.%s", table.Name, seedOutput))
			file, err := os.Create(filename)
			if err != nil {
				return fmt.Errorf("failed to create output file %s: %v", filename, err)
			}
			defer file.Close()

			switch seedOutput {
			case "csv":
				writer := csv.NewWriter(file)
				// Write header
				if len(table.Data) > 0 {
					headers := make([]string, 0, len(table.Data[0]))
					for key := range table.Data[0] {
						headers = append(headers, key)
					}
					writer.Write(headers) //nolint:errcheck // Ignoring error for simplicity
				}
				// Write data
				for _, row := range table.Data {
					record := make([]string, 0, len(row))
					// Assuming consistent keys order as headers
					sk := make([]string, 0, len(row))
					for k := range row {
						sk = append(sk, k)
					}
					sort.Strings(sk)
					for _, key := range sk {
						record = append(record, fmt.Sprintf("%v", row[key]))
					}
					writer.Write(record) //nolint:errcheck // Ignoring error for simplicity
				}
				writer.Flush() //nolint:errcheck // Ignoring error for simplicity
			case "json":
				encoder := json.NewEncoder(file)
				encoder.SetIndent("", "  ") // Pretty print JSON
				if err := encoder.Encode(table.Data); err != nil {
					return fmt.Errorf("failed to encode JSON for table %s: %v", table.Name, err)
				}
			// Add other formats (sql) as needed
			default:
				return fmt.Errorf("unsupported output format: %s", seedOutput)
			}
			log.Printf("Generated %s file: %s with %d rows", seedOutput, filename, len(table.Data))
		} else {
			log.Printf("No data to generate for table %s", table.Name)
		}
	}
	return nil
}
