package config

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"github.com/valentinesamuel/mockcraft/internal/generators"
	"gopkg.in/yaml.v3"
)

// LoadSchema loads a schema from a YAML file
func LoadSchema(path string) (*types.Schema, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	// Parse YAML into a Schema object
	var schema types.Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("failed to parse schema file: %w", err)
	}

	// Validate and enhance schema
	if err := ValidateAndEnhanceSchema(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}

// ValidateAndEnhanceSchema validates the schema and sets defaults
func ValidateAndEnhanceSchema(schema *types.Schema) error {
	if len(schema.Tables) == 0 {
		return fmt.Errorf("schema must contain at least one table")
	}

	// Track table names for relation validation
	tableNames := make(map[string]bool)
	engine := generators.GetGlobalEngine()

	for _, table := range schema.Tables {
		if table.Name == "" {
			return fmt.Errorf("table name cannot be empty")
		}
		if table.Count <= 0 {
			return fmt.Errorf("table %s must have a positive count", table.Name)
		}
		if len(table.Columns) == 0 {
			return fmt.Errorf("table %s must have at least one column", table.Name)
		}

		tableNames[table.Name] = true

		// Track column names for index validation
		columnNames := make(map[string]bool)
		for i := range table.Columns {
			col := &table.Columns[i]
			if col.Name == "" {
				return fmt.Errorf("column name cannot be empty in table %s", table.Name)
			}
			if col.Type == "" {
				return fmt.Errorf("column type cannot be empty for column %s in table %s", col.Name, table.Name)
			}

			columnNames[col.Name] = true

			// Set default industry to "base" if not specified
			if col.Industry == "" {
				col.Industry = "base"
			}

			// Set default generator based on column type if not specified
			if col.Generator == "" {
				switch col.Type {
				case "string", "text", "varchar", "char":
					col.Generator = "word"
				case "integer", "int", "bigint", "smallint":
					col.Generator = "number"
					if col.Params == nil {
						col.Params = map[string]interface{}{
							"min": 0,
							"max": 100,
						}
					}
				case "float", "decimal", "numeric":
					col.Generator = "float"
					if col.Params == nil {
						col.Params = map[string]interface{}{
							"min":       0.0,
							"max":       100.0,
							"precision": 2,
						}
					}
				case "boolean", "bool":
					col.Generator = "boolean"
				case "datetime", "timestamp":
					col.Generator = "datetime"
					if col.Params == nil {
						col.Params = map[string]interface{}{
							"format": "2006-01-02 15:04:05",
						}
					}
				case "date":
					col.Generator = "date"
					if col.Params == nil {
						col.Params = map[string]interface{}{
							"format": "2006-01-02",
						}
					}
				default:
					col.Generator = "word"
				}
			}

			// Handle enum generator values
			if col.Generator == "enum" && len(col.Values) > 0 {
				if col.Params == nil {
					col.Params = make(map[string]interface{})
				}
				col.Params["values"] = col.Values
			}

			// Validate industry and generator using the unified engine
			if err := engine.ValidateGenerator(col.Industry, col.Generator); err != nil {
				return fmt.Errorf("invalid industry '%s' or generator '%s' for column %s in table %s: %w", col.Industry, col.Generator, col.Name, table.Name, err)
			}

			// Try to generate a test value to ensure the generator works
			_, err := engine.Generate(col.Industry, col.Generator, col.Params)
			if err != nil {
				return fmt.Errorf("failed to generate test value for generator '%s' in industry '%s' for column %s in table %s: %w", col.Generator, col.Industry, col.Name, table.Name, err)
			}
		}

		// Validate indexes
		for _, index := range table.Indexes {
			if index.Name == "" {
				return fmt.Errorf("index name cannot be empty in table %s", table.Name)
			}
			if len(index.Columns) == 0 {
				return fmt.Errorf("index %s must reference at least one column in table %s", index.Name, table.Name)
			}
			for _, colName := range index.Columns {
				if !columnNames[colName] {
					return fmt.Errorf("index %s references non-existent column %s in table %s", index.Name, colName, table.Name)
				}
			}
		}
	}

	// Validate relations with enhanced foreign key checking
	validRelations := 0
	invalidRelations := []string{}
	
	for _, rel := range schema.Relations {
		if rel.Type == "" {
			errMsg := "relation type cannot be empty"
			log.Printf("ERROR: Invalid foreign key relationship - %s", errMsg)
			invalidRelations = append(invalidRelations, errMsg)
			continue
		}
		if rel.FromTable == "" {
			errMsg := "from_table cannot be empty in relation"
			log.Printf("ERROR: Invalid foreign key relationship - %s", errMsg)
			invalidRelations = append(invalidRelations, errMsg)
			continue
		}
		if rel.FromColumn == "" {
			errMsg := "from_column cannot be empty in relation"
			log.Printf("ERROR: Invalid foreign key relationship - %s", errMsg)
			invalidRelations = append(invalidRelations, errMsg)
			continue
		}
		if rel.ToTable == "" {
			errMsg := "to_table cannot be empty in relation"
			log.Printf("ERROR: Invalid foreign key relationship - %s", errMsg)
			invalidRelations = append(invalidRelations, errMsg)
			continue
		}
		if rel.ToColumn == "" {
			errMsg := "to_column cannot be empty in relation"
			log.Printf("ERROR: Invalid foreign key relationship - %s", errMsg)
			invalidRelations = append(invalidRelations, errMsg)
			continue
		}

		if !tableNames[rel.FromTable] {
			errMsg := fmt.Sprintf("relation references non-existent table %s", rel.FromTable)
			log.Printf("ERROR: Invalid foreign key relationship - %s", errMsg)
			invalidRelations = append(invalidRelations, errMsg)
			continue
		}
		if !tableNames[rel.ToTable] {
			errMsg := fmt.Sprintf("relation references non-existent table %s", rel.ToTable)
			log.Printf("ERROR: Invalid foreign key relationship - %s", errMsg)
			invalidRelations = append(invalidRelations, errMsg)
			continue
		}

		// Validate that the referenced column exists in the referenced table
		fromTableExists := false
		toTableExists := false
		fromColumnExists := false
		toColumnExists := false
		
		for _, table := range schema.Tables {
			if table.Name == rel.FromTable {
				fromTableExists = true
				for _, col := range table.Columns {
					if col.Name == rel.FromColumn {
						fromColumnExists = true
						break
					}
				}
			}
			if table.Name == rel.ToTable {
				toTableExists = true
				for _, col := range table.Columns {
					if col.Name == rel.ToColumn {
						toColumnExists = true
						break
					}
				}
			}
		}

		if !fromTableExists || !toTableExists {
			errMsg := fmt.Sprintf("relation references non-existent table(s): from=%s, to=%s", rel.FromTable, rel.ToTable)
			log.Printf("ERROR: Invalid foreign key relationship - %s", errMsg)
			invalidRelations = append(invalidRelations, errMsg)
			continue
		}

		if !fromColumnExists {
			errMsg := fmt.Sprintf("relation references non-existent column %s in table %s", rel.FromColumn, rel.FromTable)
			log.Printf("ERROR: Invalid foreign key relationship - %s", errMsg)
			invalidRelations = append(invalidRelations, errMsg)
			continue
		}

		if !toColumnExists {
			errMsg := fmt.Sprintf("relation references non-existent column %s in table %s", rel.ToColumn, rel.ToTable)
			log.Printf("ERROR: Invalid foreign key relationship - %s", errMsg)
			invalidRelations = append(invalidRelations, errMsg)
			continue
		}

		// If we reach here, the relation is valid
		validRelations++
		log.Printf("INFO: Valid foreign key relationship found: %s.%s -> %s.%s", 
			rel.FromTable, rel.FromColumn, rel.ToTable, rel.ToColumn)
	}

	// Log summary of relation validation
	log.Printf("INFO: Foreign key validation summary - Valid: %d, Invalid: %d", validRelations, len(invalidRelations))
	
	// Check minimum valid relations requirement
	if len(schema.Relations) > 0 && validRelations < 3 {
		log.Printf("WARNING: Schema has only %d valid foreign key relationships. Minimum recommended is 3.", validRelations)
		return fmt.Errorf("insufficient valid foreign key relationships: found %d, minimum required is 3", validRelations)
	}

	// If there are invalid relations, return error with details
	if len(invalidRelations) > 0 {
		log.Printf("ERROR: Schema validation failed due to %d invalid foreign key relationships", len(invalidRelations))
		return fmt.Errorf("schema has %d invalid foreign key relationships: %v", len(invalidRelations), invalidRelations)
	}

	return nil
}

// GetForeignKeyInfo returns the referenced table and column if the given column is a foreign key
func GetForeignKeyInfo(schema *types.Schema, tableName, columnName string) (string, string, bool) {
	for _, rel := range schema.Relations {
		if rel.ToTable == tableName && rel.ToColumn == columnName {
			return rel.FromTable, rel.FromColumn, true
		}
	}
	return "", "", false
}

// GenerateSchemaData generates mock data for the entire schema using the unified engine with foreign key awareness and balanced distribution
func GenerateSchemaData(schema *types.Schema) (map[string][]map[string]interface{}, error) {
	engine := generators.GetGlobalEngine()
	tableData := make(map[string][]map[string]interface{})
	primaryKeys := make(map[string][]interface{}) // table_name -> array of primary key values
	
	// Determine table dependency order
	tableOrder, err := getTableDependencyOrder(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to determine table dependency order: %w", err)
	}
	
	log.Printf("Processing tables in dependency order: %v", tableOrder)

	// Process tables in dependency order (parent tables first)
	for _, tableName := range tableOrder {
		table := findTableByName(schema, tableName)
		if table == nil {
			return nil, fmt.Errorf("table %s not found in schema", tableName)
		}
		
		if table.Count <= 0 {
			continue // Skip tables with no data to generate
		}
		
		// Convert table columns to ColumnSpecs for the unified engine
		columnSpecs := make([]generators.ColumnSpec, len(table.Columns))
		for i, col := range table.Columns {
			columnSpecs[i] = generators.ColumnSpec{
				Name:      col.Name,
				Industry:  col.Industry,
				Generator: col.Generator,
				Params:    col.Params,
			}
		}

		// Generate data for this table with foreign key awareness and balanced distribution
		data := make([]map[string]interface{}, table.Count)
		
		// Check if this table has foreign key relationships for balanced distribution
		hasForeignKeys := false
		var foreignKeyColumns []string
		var referencedTables []string
		
		for _, col := range table.Columns {
			if referencedTable, _, isForeignKey := GetForeignKeyInfo(schema, table.Name, col.Name); isForeignKey {
				hasForeignKeys = true
				foreignKeyColumns = append(foreignKeyColumns, col.Name)
				// Check if this referenced table is not already in the list
				found := false
				for _, existing := range referencedTables {
					if existing == referencedTable {
						found = true
						break
					}
				}
				if !found {
					referencedTables = append(referencedTables, referencedTable)
				}
			}
		}
		
		if hasForeignKeys && len(referencedTables) > 0 {
			// Use balanced distribution for tables with foreign keys
			log.Printf("Using balanced distribution for table %s (foreign keys found: %v)", table.Name, foreignKeyColumns)
			data = generateBalancedDistribution(engine, columnSpecs, table, schema, primaryKeys, referencedTables)
		} else {
			// Use standard generation for parent tables (no foreign keys)
			log.Printf("Using standard generation for table %s (no foreign keys)", table.Name)
			for i := 0; i < table.Count; i++ {
				row, err := engine.GenerateRow(columnSpecs, make(map[string]interface{}))
				if err != nil {
					return nil, fmt.Errorf("failed to generate data for table %s: %w", table.Name, err)
				}
				data[i] = row
			}
		}
		
		// Store primary keys for foreign key references
		primaryKeyColumn := getPrimaryKeyColumn(table)
		if primaryKeyColumn != "" {
			keys := make([]interface{}, len(data))
			for i, row := range data {
				keys[i] = row[primaryKeyColumn]
			}
			primaryKeys[table.Name] = keys
			log.Printf("Stored %d primary keys for table %s", len(keys), table.Name)
		}
		
		tableData[table.Name] = data
		log.Printf("Generated %d rows for table %s with balanced distribution", len(data), table.Name)
	}

	return tableData, nil
}

// generateBalancedDistribution creates data with balanced foreign key distribution ensuring minimum records per referenced entity
func generateBalancedDistribution(engine *generators.Engine, columnSpecs []generators.ColumnSpec, table *types.Table, schema *types.Schema, primaryKeys map[string][]interface{}, referencedTables []string) []map[string]interface{} {
	data := make([]map[string]interface{}, table.Count)
	
	// Calculate minimum records per foreign key reference
	minRecordsPerReference := 2 // Minimum 2 records per profile/entity
	
	// Find the primary referenced table (usually the first one, e.g., profiles)
	primaryReferencedTable := referencedTables[0]
	availablePrimaryKeys, exists := primaryKeys[primaryReferencedTable]
	if !exists || len(availablePrimaryKeys) == 0 {
		log.Printf("WARNING: No primary keys available for balanced distribution in table %s", table.Name)
		// Fallback to random generation
		for i := 0; i < table.Count; i++ {
			row, _ := engine.GenerateRow(columnSpecs, make(map[string]interface{}))
			data[i] = row
		}
		return data
	}
	
	numPrimaryKeys := len(availablePrimaryKeys)
	totalRecordsToGenerate := table.Count
	
	// Calculate distribution strategy
	guaranteedRecords := numPrimaryKeys * minRecordsPerReference
	remainingRecords := totalRecordsToGenerate - guaranteedRecords
	
	if remainingRecords < 0 {
		// Not enough records to guarantee minimum, distribute evenly
		log.Printf("INFO: Distributing %d records evenly across %d entities in table %s", totalRecordsToGenerate, numPrimaryKeys, table.Name)
		return generateEvenDistribution(engine, columnSpecs, table, schema, primaryKeys, totalRecordsToGenerate, availablePrimaryKeys)
	}
	
	log.Printf("INFO: Balanced distribution for table %s: %d guaranteed + %d additional records across %d entities", 
		table.Name, guaranteedRecords, remainingRecords, numPrimaryKeys)
	
	recordIndex := 0
	
	// Phase 1: Guarantee minimum records per primary key
	for _, primaryKey := range availablePrimaryKeys {
		for i := 0; i < minRecordsPerReference && recordIndex < totalRecordsToGenerate; i++ {
			row, err := engine.GenerateRow(columnSpecs, make(map[string]interface{}))
			if err != nil {
				log.Printf("ERROR: Failed to generate row: %v", err)
				continue
			}
			
			// Set foreign key values
			setForeignKeyValues(row, table, schema, primaryKeys, primaryKey, primaryReferencedTable)
			data[recordIndex] = row
			recordIndex++
		}
	}
	
	// Phase 2: Distribute remaining records with weighted random selection
	usageCounts := make(map[interface{}]int)
	for _, key := range availablePrimaryKeys {
		usageCounts[key] = minRecordsPerReference
	}
	
	for recordIndex < totalRecordsToGenerate {
		// Select primary key with weighted preference for less used ones
		selectedKey := selectWeightedRandomKey(availablePrimaryKeys, usageCounts)
		
		row, err := engine.GenerateRow(columnSpecs, make(map[string]interface{}))
		if err != nil {
			log.Printf("ERROR: Failed to generate row: %v", err)
			recordIndex++
			continue
		}
		
		// Set foreign key values
		setForeignKeyValues(row, table, schema, primaryKeys, selectedKey, primaryReferencedTable)
		data[recordIndex] = row
		usageCounts[selectedKey]++
		recordIndex++
	}
	
	// Log distribution summary
	logDistributionSummary(table.Name, usageCounts)
	
	return data
}

// generateEvenDistribution distributes records as evenly as possible across available primary keys
func generateEvenDistribution(engine *generators.Engine, columnSpecs []generators.ColumnSpec, table *types.Table, schema *types.Schema, primaryKeys map[string][]interface{}, totalRecords int, availablePrimaryKeys []interface{}) []map[string]interface{} {
	data := make([]map[string]interface{}, totalRecords)
	recordsPerKey := totalRecords / len(availablePrimaryKeys)
	extraRecords := totalRecords % len(availablePrimaryKeys)
	
	recordIndex := 0
	primaryReferencedTable := ""
	
	// Find the primary referenced table name
	for tableName, keys := range primaryKeys {
		if len(keys) > 0 && keys[0] == availablePrimaryKeys[0] {
			primaryReferencedTable = tableName
			break
		}
	}
	
	for keyIndex, primaryKey := range availablePrimaryKeys {
		recordsForThisKey := recordsPerKey
		if keyIndex < extraRecords {
			recordsForThisKey++ // Distribute extra records
		}
		
		for i := 0; i < recordsForThisKey && recordIndex < totalRecords; i++ {
			row, err := engine.GenerateRow(columnSpecs, make(map[string]interface{}))
			if err != nil {
				log.Printf("ERROR: Failed to generate row: %v", err)
				recordIndex++
				continue
			}
			
			setForeignKeyValues(row, table, schema, primaryKeys, primaryKey, primaryReferencedTable)
			data[recordIndex] = row
			recordIndex++
		}
	}
	
	return data
}

// setForeignKeyValues sets all foreign key values for a row
func setForeignKeyValues(row map[string]interface{}, table *types.Table, schema *types.Schema, primaryKeys map[string][]interface{}, primaryKey interface{}, primaryReferencedTable string) {
	for _, col := range table.Columns {
		if referencedTable, referencedColumn, isForeignKey := GetForeignKeyInfo(schema, table.Name, col.Name); isForeignKey {
			if referencedTable == primaryReferencedTable {
				// Use the assigned primary key for the main referenced table
				row[col.Name] = primaryKey
				log.Printf("Set primary foreign key %s.%s = %v (referencing %s.%s)", table.Name, col.Name, primaryKey, referencedTable, referencedColumn)
			} else {
				// For other foreign key references, use random selection
				if availableKeys, exists := primaryKeys[referencedTable]; exists && len(availableKeys) > 0 {
					randomKey := availableKeys[rand.Intn(len(availableKeys))]
					row[col.Name] = randomKey
					log.Printf("Set secondary foreign key %s.%s = %v (referencing %s.%s)", table.Name, col.Name, randomKey, referencedTable, referencedColumn)
				}
			}
		}
	}
}

// selectWeightedRandomKey selects a key with preference for less frequently used ones
func selectWeightedRandomKey(availableKeys []interface{}, usageCounts map[interface{}]int) interface{} {
	if len(availableKeys) == 0 {
		return nil
	}
	
	// Find minimum usage count
	minUsage := usageCounts[availableKeys[0]]
	for _, key := range availableKeys {
		if usageCounts[key] < minUsage {
			minUsage = usageCounts[key]
		}
	}
	
	// Collect keys with minimum usage
	minUsageKeys := []interface{}{}
	for _, key := range availableKeys {
		if usageCounts[key] == minUsage {
			minUsageKeys = append(minUsageKeys, key)
		}
	}
	
	// Randomly select from minimum usage keys
	return minUsageKeys[rand.Intn(len(minUsageKeys))]
}

// logDistributionSummary logs how records were distributed across foreign keys
func logDistributionSummary(tableName string, usageCounts map[interface{}]int) {
	log.Printf("=== Distribution Summary for %s ===", tableName)
	for key, count := range usageCounts {
		log.Printf("  %v: %d records", key, count)
	}
	log.Printf("=== End Distribution Summary ===")
}

// getTableDependencyOrder returns tables in dependency order (parent tables first)
func getTableDependencyOrder(schema *types.Schema) ([]string, error) {
	// Create dependency graph
	dependencies := make(map[string][]string) // child -> parents
	allTables := make(map[string]bool)
	
	// Initialize all tables
	for _, table := range schema.Tables {
		allTables[table.Name] = true
		dependencies[table.Name] = []string{}
	}
	
	// Build dependency relationships
	// In a foreign key relationship, the ToTable (child) depends on the FromTable (parent)
	for _, rel := range schema.Relations {
		if rel.FromTable != rel.ToTable { // Avoid self-references
			dependencies[rel.ToTable] = append(dependencies[rel.ToTable], rel.FromTable)
		}
	}
	
	// Debug: print the dependency graph
	for table, deps := range dependencies {
		log.Printf("Table %s depends on: %v", table, deps)
	}
	
	// Topological sort using Kahn's algorithm
	var result []string
	inDegree := make(map[string]int)
	
	// Calculate in-degrees (number of dependencies each table has)
	for table := range allTables {
		inDegree[table] = len(dependencies[table])
	}
	
	// Find tables with no dependencies (in-degree 0)
	queue := []string{}
	for table, degree := range inDegree {
		log.Printf("Table %s has in-degree: %d", table, degree)
		if degree == 0 {
			queue = append(queue, table)
		}
	}
	log.Printf("Initial queue (0 in-degree): %v", queue)
	
	// Process queue
	for len(queue) > 0 {
		// Remove first element
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		log.Printf("Processing table: %s", current)
		
		// Reduce in-degree for tables that depend on current table
		for table, deps := range dependencies {
			for _, dep := range deps {
				if dep == current {
					inDegree[table]--
					log.Printf("Reduced in-degree of %s to %d", table, inDegree[table])
					if inDegree[table] == 0 {
						queue = append(queue, table)
						log.Printf("Added %s to queue", table)
					}
				}
			}
		}
	}
	
	// Check for circular dependencies
	if len(result) != len(allTables) {
		return nil, fmt.Errorf("circular dependency detected in table relationships")
	}
	
	// Result already has parent tables first due to corrected dependency logic
	return result, nil
}

// findTableByName finds a table by name in the schema
func findTableByName(schema *types.Schema, tableName string) *types.Table {
	for i := range schema.Tables {
		if schema.Tables[i].Name == tableName {
			return &schema.Tables[i]
		}
	}
	return nil
}

// getPrimaryKeyColumn returns the name of the primary key column for a table
func getPrimaryKeyColumn(table *types.Table) string {
	for _, col := range table.Columns {
		if col.IsPrimary {
			return col.Name
		}
	}
	return ""
}
