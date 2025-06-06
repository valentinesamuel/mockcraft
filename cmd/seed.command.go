package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/valentinesamuel/mockcraft/internal/config"
	"github.com/valentinesamuel/mockcraft/internal/database"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"github.com/valentinesamuel/mockcraft/internal/generators"
	"github.com/valentinesamuel/mockcraft/internal/output"
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

func topologicalSortTables(tables []types.Table, relations []types.Relationship) []types.Table {
	tableMap := make(map[string]types.Table)
	for _, t := range tables {
		tableMap[t.Name] = t
	}

	// Build adjacency list
	adj := make(map[string][]string)
	inDegree := make(map[string]int)
	for _, t := range tables {
		adj[t.Name] = []string{}
		inDegree[t.Name] = 0
	}
	for _, rel := range relations {
		adj[rel.FromTable] = append(adj[rel.FromTable], rel.ToTable)
		inDegree[rel.ToTable]++
	}

	// Kahn's algorithm
	var sorted []string
	var queue []string
	for name, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, name)
		}
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		sorted = append(sorted, n)
		for _, m := range adj[n] {
			inDegree[m]--
			if inDegree[m] == 0 {
				queue = append(queue, m)
			}
		}
	}

	// Build sorted tables
	var result []types.Table
	for _, name := range sorted {
		result = append(result, tableMap[name])
	}
	return result
}

// verifyReferentialIntegrity checks that all foreign key references are valid
func verifyReferentialIntegrity(ctx context.Context, db types.Database, schema *types.Schema) error {
	log.Println("Verifying referential integrity...")

	// For each relationship in the schema
	for _, rel := range schema.Relations {
		// Skip if any table or column name is empty
		if rel.FromTable == "" || rel.FromColumn == "" || rel.ToTable == "" || rel.ToColumn == "" {
			continue
		}
		log.Printf("Checking relationship: %s.%s -> %s.%s", rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn)

		// Get all IDs from the parent table
		parentIDs, err := db.GetAllIDs(ctx, rel.FromTable)
		if err != nil {
			return fmt.Errorf("failed to get IDs from parent table %s: %w", rel.FromTable, err)
		}

		// Create a map for O(1) lookup
		parentIDMap := make(map[interface{}]bool)
		for _, id := range parentIDs {
			parentIDMap[id] = true
		}

		// Get all foreign key values from the child table
		childFKs, err := db.GetAllForeignKeys(ctx, rel.ToTable, rel.ToColumn)
		if err != nil {
			return fmt.Errorf("failed to get foreign keys from child table %s: %w", rel.ToTable, err)
		}

		// Check each foreign key value
		var invalidRefs []interface{}
		for _, fk := range childFKs {
			if !parentIDMap[fk] {
				invalidRefs = append(invalidRefs, fk)
			}
		}

		// Report any invalid references
		if len(invalidRefs) > 0 {
			return fmt.Errorf("found %d invalid references in %s.%s: %v",
				len(invalidRefs), rel.ToTable, rel.ToColumn, invalidRefs)
		}

		log.Printf("✓ Relationship %s.%s -> %s.%s is valid",
			rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn)
	}

	log.Println("All referential integrity checks passed!")
	return nil
}

// verifyUserComments checks that each user has at least one comment
func verifyUserComments(ctx context.Context, db types.Database) error {
	log.Println("\nVerifying user comments...")

	// Get all user IDs
	userIDs, err := db.GetAllIDs(ctx, "users")
	if err != nil {
		return fmt.Errorf("failed to get user IDs: %w", err)
	}

	// For each user, check if they have any comments
	for _, userID := range userIDs {
		// Get comments for this user
		comments, err := db.GetAllForeignKeys(ctx, "comments", "user_id")
		if err != nil {
			return fmt.Errorf("failed to get comments for user %v: %w", userID, err)
		}

		// Count comments for this user
		commentCount := 0
		for _, commentUserID := range comments {
			if commentUserID == userID {
				commentCount++
			}
		}

		log.Printf("User %v has %d comments", userID, commentCount)
		if commentCount == 0 {
			return fmt.Errorf("user %v has no comments", userID)
		}
	}

	log.Println("All users have at least one comment!")
	return nil
}

// verifyAllParentReferences checks that every parent record is referenced at least once in the child table for all relationships
func verifyAllParentReferences(ctx context.Context, db types.Database, schema *types.Schema) error {
	fmt.Println("[DEBUG] Starting relationship verification...")
	log.Println("[DEBUG] Starting relationship verification...")
	fmt.Printf("[DEBUG] schema.Relations: %+v\n", schema.Relations)

	// First, collect all foreign key relationships from the schema
	type Relationship struct {
		FromTable  string
		FromColumn string
		ToTable    string
		ToColumn   string
	}

	relationships := make([]Relationship, 0)
	for _, rel := range schema.Relations {
		if rel.FromTable == "" || rel.FromColumn == "" || rel.ToTable == "" || rel.ToColumn == "" {
			continue
		}
		relationships = append(relationships, Relationship{
			FromTable:  rel.FromTable,
			FromColumn: rel.FromColumn,
			ToTable:    rel.ToTable,
			ToColumn:   rel.ToColumn,
		})
	}

	// For each relationship, get all parent IDs and count their references
	for _, rel := range relationships {
		fmt.Printf("\n[DEBUG] Checking relationship: %s.%s -> %s.%s\n", rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn)
		// panic("DEBUG PANIC: verifyAllParentReferences is running!")
		// Get all parent IDs
		parentIDs, err := db.GetAllForeignKeys(ctx, rel.FromTable, rel.FromColumn)
		if err != nil {
			fmt.Printf("[DEBUG] Failed to get parent IDs from %s.%s: %v\n", rel.FromTable, rel.FromColumn, err)
			return fmt.Errorf("failed to get parent IDs from %s.%s: %w", rel.FromTable, rel.FromColumn, err)
		}
		fmt.Printf("[DEBUG] parentIDs (%s.%s): %v\n", rel.FromTable, rel.FromColumn, parentIDs)

		// Get all child foreign keys
		childFKs, err := db.GetAllForeignKeys(ctx, rel.ToTable, rel.ToColumn)
		if err != nil {
			fmt.Printf("[DEBUG] Failed to get child FKs from %s.%s: %v\n", rel.ToTable, rel.ToColumn, err)
			return fmt.Errorf("failed to get child FKs from %s.%s: %w", rel.ToTable, rel.ToColumn, err)
		}
		fmt.Printf("[DEBUG] childFKs (%s.%s): %v\n", rel.ToTable, rel.ToColumn, childFKs)

		// Count references for each parent ID
		fkCount := make(map[interface{}]int)
		for _, fk := range childFKs {
			fkCount[fk]++
		}

		// Log the count for each parent
		totalReferences := 0
		for _, pid := range parentIDs {
			count := fkCount[pid]
			totalReferences += count
			log.Printf("There are %d %s from %s %v", count, rel.ToTable, rel.FromTable, pid)
			fmt.Printf("There are %d %s from %s %v\n", count, rel.ToTable, rel.FromTable, pid)
		}

		// Log summary
		log.Printf("Total %s records: %d", rel.FromTable, len(parentIDs))
		log.Printf("Total %s records: %d", rel.ToTable, len(childFKs))
		log.Printf("Average %s per %s: %.2f", rel.ToTable, rel.FromTable, float64(totalReferences)/float64(len(parentIDs)))
	}

	return nil
}

func seedDatabase(schema *types.Schema) error {
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

	// Topologically sort tables by dependencies
	sortedTables := topologicalSortTables(schema.Tables, schema.Relations)

	// Store generated IDs for foreign key references
	tableIDs := make(map[string][]interface{})

	var userIDs []interface{}
	var commentData []map[string]interface{}
	var commentTable *types.Table

	// Process each table in sorted order
	for _, table := range sortedTables {
		if seedDryRun {
			log.Printf("Would seed table %s with %d rows", table.Name, table.Count)
			continue
		}

		// Create table
		if err := db.CreateTable(ctx, table.Name, table.Columns); err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.Name, err)
		}

		// Create indexes
		for _, index := range table.Indexes {
			if err := db.CreateIndex(ctx, table.Name, index); err != nil {
				return fmt.Errorf("failed to create index %s on table %s: %w", index.Name, table.Name, err)
			}
		}

		if table.Name == "comments" {
			commentTable = &table
			continue // Defer comment generation until after users are processed
		}

		// Generate data
		table.Data = make([]map[string]interface{}, table.Count)
		for i := 0; i < table.Count; i++ {
			row := make(map[string]interface{})
			for _, col := range table.Columns {
				// Handle foreign key fields
				if strings.HasSuffix(col.Name, "_id") {
					var referencedTable string
					for _, rel := range schema.Relations {
						if rel.ToColumn == col.Name {
							referencedTable = rel.FromTable
							break
						}
					}
					if ids, ok := tableIDs[referencedTable]; ok && len(ids) > 0 {
						row[col.Name] = ids[rand.Intn(len(ids))]
						continue
					}
				}

				// Get the appropriate generator based on column type
				generator, err := generators.Get(col.Generator)
				if err != nil {
					// If no specific generator found, try to infer from column name
					switch {
					case strings.Contains(strings.ToLower(col.Name), "email"):
						generator, _ = generators.Get("email")
					case strings.Contains(strings.ToLower(col.Name), "name"):
						generator, _ = generators.Get("firstname")
					case strings.Contains(strings.ToLower(col.Name), "address"):
						generator, _ = generators.Get("address")
					case strings.Contains(strings.ToLower(col.Name), "phone"):
						generator, _ = generators.Get("phone")
					case strings.Contains(strings.ToLower(col.Name), "company"):
						generator, _ = generators.Get("company")
					case strings.Contains(strings.ToLower(col.Name), "job"):
						generator, _ = generators.Get("job_title")
					case strings.Contains(strings.ToLower(col.Name), "url"):
						generator, _ = generators.Get("url")
					case strings.Contains(strings.ToLower(col.Name), "ip"):
						generator, _ = generators.Get("ip")
					case strings.Contains(strings.ToLower(col.Name), "domain"):
						generator, _ = generators.Get("domain")
					case strings.Contains(strings.ToLower(col.Name), "username"):
						generator, _ = generators.Get("username")
					case strings.Contains(strings.ToLower(col.Name), "credit"):
						generator, _ = generators.Get("credit_card")
					case strings.Contains(strings.ToLower(col.Name), "currency"):
						generator, _ = generators.Get("currency")
					case strings.Contains(strings.ToLower(col.Name), "price"):
						generator, _ = generators.Get("price")
					case strings.Contains(strings.ToLower(col.Name), "product"):
						generator, _ = generators.Get("product")
					default:
						// Fallback to basic generators based on type
						switch col.Type {
						case "string":
							generator, _ = generators.Get("text")
						case "integer":
							generator, _ = generators.Get("number")
						case "decimal":
							generator, _ = generators.Get("decimal")
						case "boolean":
							generator, _ = generators.Get("boolean")
						case "timestamp":
							generator, _ = generators.Get("timestamp")
						}
					}
				}

				// Generate the value with parameters if available
				if generator != nil {
					value, err := generator.Generate(col.Params)
					if err == nil {
						row[col.Name] = value
					} else {
						row[col.Name] = fmt.Sprintf("error-%d", i)
					}
				} else {
					row[col.Name] = fmt.Sprintf("unknown-%d", i)
				}
			}
			if id, ok := row["_id"]; ok {
				tableIDs[table.Name] = append(tableIDs[table.Name], id)
				if table.Name == "users" {
					userIDs = append(userIDs, id)
				}
			}
			table.Data[i] = row
		}
		if len(table.Data) > 0 {
			if err := db.InsertData(ctx, table.Name, table.Data); err != nil {
				return fmt.Errorf("failed to insert data into table %s: %w", table.Name, err)
			}
		}
		log.Printf("Seeded table %s with %d rows", table.Name, len(table.Data))
	}

	// Now generate comments, ensuring each user gets at least one, and all columns are generated
	if commentTable != nil {
		commentData = make([]map[string]interface{}, commentTable.Count)
		// First, assign one comment to each user
		for i := 0; i < len(userIDs) && i < commentTable.Count; i++ {
			row := make(map[string]interface{})
			for _, col := range commentTable.Columns {
				if col.Name == "user_id" {
					row[col.Name] = userIDs[i]
					continue
				}
				if strings.HasSuffix(col.Name, "_id") {
					var referencedTable string
					for _, rel := range schema.Relations {
						if rel.ToColumn == col.Name {
							referencedTable = rel.FromTable
							break
						}
					}
					if ids, ok := tableIDs[referencedTable]; ok && len(ids) > 0 {
						row[col.Name] = ids[rand.Intn(len(ids))]
						continue
					}
				}
				generator, err := generators.Get(col.Generator)
				if err != nil {
					// If no specific generator found, try to infer from column name
					switch {
					case strings.Contains(strings.ToLower(col.Name), "email"):
						generator, _ = generators.Get("email")
					case strings.Contains(strings.ToLower(col.Name), "name"):
						generator, _ = generators.Get("firstname")
					case strings.Contains(strings.ToLower(col.Name), "address"):
						generator, _ = generators.Get("address")
					case strings.Contains(strings.ToLower(col.Name), "phone"):
						generator, _ = generators.Get("phone")
					case strings.Contains(strings.ToLower(col.Name), "company"):
						generator, _ = generators.Get("company")
					case strings.Contains(strings.ToLower(col.Name), "job"):
						generator, _ = generators.Get("job_title")
					case strings.Contains(strings.ToLower(col.Name), "url"):
						generator, _ = generators.Get("url")
					case strings.Contains(strings.ToLower(col.Name), "ip"):
						generator, _ = generators.Get("ip")
					case strings.Contains(strings.ToLower(col.Name), "domain"):
						generator, _ = generators.Get("domain")
					case strings.Contains(strings.ToLower(col.Name), "username"):
						generator, _ = generators.Get("username")
					case strings.Contains(strings.ToLower(col.Name), "credit"):
						generator, _ = generators.Get("credit_card")
					case strings.Contains(strings.ToLower(col.Name), "currency"):
						generator, _ = generators.Get("currency")
					case strings.Contains(strings.ToLower(col.Name), "price"):
						generator, _ = generators.Get("price")
					case strings.Contains(strings.ToLower(col.Name), "product"):
						generator, _ = generators.Get("product")
					default:
						// Fallback to basic generators based on type
						switch col.Type {
						case "string":
							generator, _ = generators.Get("text")
						case "integer":
							generator, _ = generators.Get("number")
						case "decimal":
							generator, _ = generators.Get("decimal")
						case "boolean":
							generator, _ = generators.Get("boolean")
						case "timestamp":
							generator, _ = generators.Get("timestamp")
						}
					}
				}
				if generator != nil {
					value, err := generator.Generate(col.Params)
					if err == nil {
						row[col.Name] = value
					} else {
						row[col.Name] = fmt.Sprintf("error-%d", i)
					}
				} else {
					row[col.Name] = fmt.Sprintf("unknown-%d", i)
				}
			}
			commentData[i] = row
		}
		// Then, assign remaining comments randomly
		for i := len(userIDs); i < commentTable.Count; i++ {
			row := make(map[string]interface{})
			for _, col := range commentTable.Columns {
				if col.Name == "user_id" {
					row[col.Name] = userIDs[rand.Intn(len(userIDs))]
					continue
				}
				if strings.HasSuffix(col.Name, "_id") {
					var referencedTable string
					for _, rel := range schema.Relations {
						if rel.ToColumn == col.Name {
							referencedTable = rel.FromTable
							break
						}
					}
					if ids, ok := tableIDs[referencedTable]; ok && len(ids) > 0 {
						row[col.Name] = ids[rand.Intn(len(ids))]
						continue
					}
				}
				generator, err := generators.Get(col.Generator)
				if err != nil {
					// If no specific generator found, try to infer from column name
					switch {
					case strings.Contains(strings.ToLower(col.Name), "email"):
						generator, _ = generators.Get("email")
					case strings.Contains(strings.ToLower(col.Name), "name"):
						generator, _ = generators.Get("firstname")
					case strings.Contains(strings.ToLower(col.Name), "address"):
						generator, _ = generators.Get("address")
					case strings.Contains(strings.ToLower(col.Name), "phone"):
						generator, _ = generators.Get("phone")
					case strings.Contains(strings.ToLower(col.Name), "company"):
						generator, _ = generators.Get("company")
					case strings.Contains(strings.ToLower(col.Name), "job"):
						generator, _ = generators.Get("job_title")
					case strings.Contains(strings.ToLower(col.Name), "url"):
						generator, _ = generators.Get("url")
					case strings.Contains(strings.ToLower(col.Name), "ip"):
						generator, _ = generators.Get("ip")
					case strings.Contains(strings.ToLower(col.Name), "domain"):
						generator, _ = generators.Get("domain")
					case strings.Contains(strings.ToLower(col.Name), "username"):
						generator, _ = generators.Get("username")
					case strings.Contains(strings.ToLower(col.Name), "credit"):
						generator, _ = generators.Get("credit_card")
					case strings.Contains(strings.ToLower(col.Name), "currency"):
						generator, _ = generators.Get("currency")
					case strings.Contains(strings.ToLower(col.Name), "price"):
						generator, _ = generators.Get("price")
					case strings.Contains(strings.ToLower(col.Name), "product"):
						generator, _ = generators.Get("product")
					default:
						// Fallback to basic generators based on type
						switch col.Type {
						case "string":
							generator, _ = generators.Get("text")
						case "integer":
							generator, _ = generators.Get("number")
						case "decimal":
							generator, _ = generators.Get("decimal")
						case "boolean":
							generator, _ = generators.Get("boolean")
						case "timestamp":
							generator, _ = generators.Get("timestamp")
						}
					}
				}
				if generator != nil {
					value, err := generator.Generate(col.Params)
					if err == nil {
						row[col.Name] = value
					} else {
						row[col.Name] = fmt.Sprintf("error-%d", i)
					}
				} else {
					row[col.Name] = fmt.Sprintf("unknown-%d", i)
				}
			}
			commentData[i] = row
		}
		if len(commentData) > 0 {
			if err := db.InsertData(ctx, "comments", commentData); err != nil {
				return fmt.Errorf("failed to insert data into table comments: %w", err)
			}
		}
		log.Printf("Seeded table comments with %d rows", len(commentData))
	}

	// Verify referential integrity after seeding
	if err := verifyReferentialIntegrity(ctx, db, schema); err != nil {
		return fmt.Errorf("referential integrity check failed: %w", err)
	}

	fmt.Println("[DEBUG] About to call verifyAllParentReferences...")
	if err := verifyAllParentReferences(ctx, db, schema); err != nil {
		return fmt.Errorf("parent reference verification failed: %w", err)
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

		// Generate data if not already present
		if len(table.Data) == 0 {
			// Use count from flag if specified, otherwise use table's count
			count := seedCount
			if count == 0 {
				count = table.Count
			}
			if count == 0 {
				count = 100 // Default to 100 rows if no count specified
			}

			// Generate data for each row
			table.Data = make([]map[string]interface{}, count)
			for i := 0; i < count; i++ {
				row := make(map[string]interface{})
				for _, col := range table.Columns {
					// Get the appropriate generator based on column type
					generator, err := generators.Get(col.Generator)
					if err != nil {
						// If no specific generator found, try to infer from column name
						switch {
						case strings.Contains(strings.ToLower(col.Name), "email"):
							generator, _ = generators.Get("email")
						case strings.Contains(strings.ToLower(col.Name), "name"):
							generator, _ = generators.Get("firstname")
						case strings.Contains(strings.ToLower(col.Name), "address"):
							generator, _ = generators.Get("address")
						case strings.Contains(strings.ToLower(col.Name), "phone"):
							generator, _ = generators.Get("phone")
						case strings.Contains(strings.ToLower(col.Name), "company"):
							generator, _ = generators.Get("company")
						case strings.Contains(strings.ToLower(col.Name), "job"):
							generator, _ = generators.Get("job_title")
						case strings.Contains(strings.ToLower(col.Name), "url"):
							generator, _ = generators.Get("url")
						case strings.Contains(strings.ToLower(col.Name), "ip"):
							generator, _ = generators.Get("ip")
						case strings.Contains(strings.ToLower(col.Name), "domain"):
							generator, _ = generators.Get("domain")
						case strings.Contains(strings.ToLower(col.Name), "username"):
							generator, _ = generators.Get("username")
						case strings.Contains(strings.ToLower(col.Name), "credit"):
							generator, _ = generators.Get("credit_card")
						case strings.Contains(strings.ToLower(col.Name), "currency"):
							generator, _ = generators.Get("currency")
						case strings.Contains(strings.ToLower(col.Name), "price"):
							generator, _ = generators.Get("price")
						case strings.Contains(strings.ToLower(col.Name), "product"):
							generator, _ = generators.Get("product")
						default:
							// Fallback to basic generators based on type
							switch col.Type {
							case "string":
								generator, _ = generators.Get("text")
							case "integer":
								generator, _ = generators.Get("number")
							case "decimal":
								generator, _ = generators.Get("decimal")
							case "boolean":
								generator, _ = generators.Get("boolean")
							case "timestamp":
								generator, _ = generators.Get("timestamp")
							}
						}
					}

					// Generate the value with parameters if available
					if generator != nil {
						value, err := generator.Generate(col.Params)
						if err == nil {
							row[col.Name] = value
						} else {
							row[col.Name] = fmt.Sprintf("error-%d", i)
						}
					} else {
						row[col.Name] = fmt.Sprintf("unknown-%d", i)
					}
				}
				table.Data[i] = row
			}
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
	return output.WriteSQL(filename, table)
}
