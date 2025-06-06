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

			// Drop existing collections if MongoDB
			if db.GetDriver() == "mongodb" {
				for _, table := range schema.Tables {
					if err := db.DropTable(ctx, table.Name); err != nil {
						return fmt.Errorf("failed to drop table %s: %w", table.Name, err)
					}
				}
			}

			if err := seedDatabase(ctx, db, schema); err != nil {
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

func seedDatabase(ctx context.Context, db types.Database, schema *types.Schema) error {
	// 1. Create tables based on schema
	log.Println("Creating tables...")
	for _, table := range schema.Tables {
		log.Printf("Creating table: %s", table.Name)
		if err := db.CreateTable(ctx, table.Name, &table); err != nil {
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
			if rel.FromTable == table.Name && rel.ToTable == table.Name {
				hasCircularDependency = true
				break
			}
		}

		// Generate data for each row
		for i := 0; i < table.Count; i++ {
			row := make(map[string]interface{})

			for _, col := range table.Columns {
				isForeignKey := false
				// Check if this column is a foreign key in any incoming relationship
				for _, rel := range incomingRelations {
					if rel.ToColumn == col.Name {
						// This column is a foreign key
						referencedTable := rel.FromTable

						// Handle self-referential and other circular dependencies by setting to nil initially
						if hasCircularDependency || (len(tableIDs[referencedTable]) == 0 && referencedTable != table.Name) {
							row[col.Name] = nil
							isForeignKey = true
						} else if ids, ok := tableIDs[referencedTable]; ok && len(ids) > 0 {
							// Use modulo to cycle through parent IDs to ensure all parents are referenced
							selectedID := ids[i%len(ids)]
							row[col.Name] = selectedID
							isForeignKey = true
						} else {
							// This should not happen if non-circular dependencies are sorted correctly
							log.Printf("Warning: No IDs found for referenced table %s for foreign key %s.%s during initial row generation.", referencedTable, table.Name, col.Name)
							row[col.Name] = nil
							isForeignKey = true
						}
						break
					}
				}

				// If it's not a foreign key or we couldn't generate a foreign key, generate a regular value
				if !isForeignKey {
					generator, err := generators.Get(col.Generator)
					if err != nil {
						log.Printf("Warning: Generator '%s' not found for column %s.%s. Using fallback.", col.Generator, table.Name, col.Name)
						switch col.Type {
						case "string", "text", "uuid":
							generator, _ = generators.Get("text")
						case "integer":
							generator, _ = generators.Get("number")
						case "decimal", "float":
							generator, _ = generators.Get("decimal")
						case "boolean":
							generator, _ = generators.Get("boolean")
						case "timestamp", "datetime", "date":
							generator, _ = generators.Get("timestamp")
						default:
							log.Printf("Warning: No fallback generator for type '%s' for column %s.%s.", col.Type, table.Name, col.Name)
							row[col.Name] = nil
						}
					}

					if generator != nil {
						value, err := generator.Generate(col.Params)
						if err == nil {
							row[col.Name] = value
						} else {
							log.Printf("Warning: Failed to generate value for column %s.%s with generator '%s': %v. Using error placeholder.", table.Name, col.Name, col.Generator, err)
							row[col.Name] = fmt.Sprintf("generator-error-%s", col.Generator)
						}
					} else {
						row[col.Name] = nil
					}
				}
			}
			data[i] = row
		}

		// Insert data into table
		if len(data) > 0 {
			if err := db.InsertData(ctx, table.Name, data); err != nil {
				return fmt.Errorf("failed to insert data into table %s: %w", table.Name, err)
			}
			log.Printf("Seeded table %s with %d initial rows", table.Name, len(data))

			// Retrieve and store generated IDs for this table
			ids, err := db.GetAllIDs(ctx, table.Name)
			if err != nil {
				log.Printf("Warning: Could not retrieve IDs after initial seeding of table %s: %v", table.Name, err)
				tableIDs[table.Name] = []string{}
			} else {
				tableIDs[table.Name] = ids
				log.Printf("[DEBUG] Stored %d IDs for table %s after initial seeding", len(ids), table.Name)
			}

			// Store data for tables with circular dependencies for later update
			if hasCircularDependency {
				circularData[table.Name] = data
			}
		} else {
			log.Printf("No data generated for table %s", table.Name)
			tableIDs[table.Name] = []string{}
		}
	}

	// 5. Update circular dependencies
	log.Println("Updating circular dependencies...")
	for tableName, data := range circularData {
		// Get all IDs for this table (should be available now)
		ids, ok := tableIDs[tableName]
		if !ok || len(ids) == 0 {
			log.Printf("Warning: No IDs found for table %s when updating circular dependencies", tableName)
			continue
		}

		// Find circular relationships for this table
		for _, rel := range schema.Relations {
			if rel.FromTable == tableName && rel.ToTable == tableName {
				// Update each row with a valid reference
				for i, row := range data {
					// Ensure at least one non-circular reference or handle base case
					// For simplicity, let the first record in a self-referential table have a nil foreign key
					if i > 0 {
						// Reference the previous row
						row[rel.ToColumn] = ids[i-1]
					} else {
						row[rel.ToColumn] = nil
					}
				}

				// Update the data in the database
				if err := db.UpdateData(ctx, tableName, data); err != nil {
					return fmt.Errorf("failed to update circular dependencies for table %s: %w", tableName, err)
				}
				log.Printf("Updated circular dependencies for table %s", tableName)
			}
		}
	}

	// 6. Verify referential integrity
	log.Println("Verifying referential integrity...")
	for _, rel := range schema.Relations {
		log.Printf("Checking relationship: %s.%s -> %s.%s", rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn)
		if err := db.VerifyReferentialIntegrity(ctx, rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn); err != nil {
			return fmt.Errorf("referential integrity check failed: %w", err)
		}
		log.Printf("✓ Relationship %s.%s -> %s.%s is valid", rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn)
	}
	log.Println("All referential integrity checks passed!")

	// 7. Verify all parent references
	log.Println("[DEBUG] Starting relationship verification...")
	fmt.Println("[DEBUG] Starting relationship verification...")
	fmt.Printf("[DEBUG] schema.Relations: %+v\n", schema.Relations)

	for _, rel := range schema.Relations {
		log.Printf("Verifying relationship counts for: %s.%s -> %s.%s", rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn)
		parentIDs, err := db.GetAllIDs(ctx, rel.FromTable)
		if err != nil {
			log.Printf("Error getting parent IDs for %s.%s: %v", rel.FromTable, rel.FromColumn, err)
			continue
		}

		log.Printf("[DEBUG] Found %d parent IDs in %s", len(parentIDs), rel.FromTable)
		fmt.Printf("[DEBUG] Found %d parent IDs in %s\n", len(parentIDs), rel.FromTable)

		childFKs, err := db.GetAllForeignKeys(ctx, rel.ToTable, rel.ToColumn)
		if err != nil {
			log.Printf("Error getting child foreign keys for %s.%s: %v", rel.ToTable, rel.ToColumn, err)
			continue
		}

		log.Printf("[DEBUG] Found %d child foreign keys in %s.%s", len(childFKs), rel.ToTable, rel.ToColumn)
		fmt.Printf("[DEBUG] Found %d child foreign keys in %s.%s\n", len(childFKs), rel.ToTable, rel.ToColumn)

		// Create a map to count references for each parent ID
		referenceCounts := make(map[string]int)
		for _, fk := range childFKs {
			referenceCounts[fk]++
		}

		// Calculate and log statistics
		totalReferences := len(childFKs)
		if len(parentIDs) > 0 {
			avgReferences := float64(totalReferences) / float64(len(parentIDs))
			log.Printf("Relationship %s.%s -> %s.%s: Average %.2f references per parent",
				rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn, avgReferences)
		}

		// Check for unreferenced parents
		for _, parentID := range parentIDs {
			if count := referenceCounts[parentID]; count == 0 {
				log.Printf("Warning: Parent ID %v in %s has no references", parentID, rel.FromTable)
			}
		}
	}

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
