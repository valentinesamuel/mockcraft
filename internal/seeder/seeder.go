package seeder

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/valentinesamuel/mockcraft/internal/config"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"github.com/valentinesamuel/mockcraft/internal/generators"
)

// Seeder represents a database seeder
type Seeder struct {
	db     types.Database
	engine *generators.Engine
}

// New creates a new database seeder
func New(db types.Database) *Seeder {
	return &Seeder{
		db:     db,
		engine: generators.GetGlobalEngine(),
	}
}

// Seed seeds the database with mock data
func (s *Seeder) Seed(ctx context.Context, schema *types.Schema) error {
	// Create tables
	for _, table := range schema.Tables {
		// Validate industry and generator for each column
		for _, column := range table.Columns {
			if column.Industry != "" && column.Generator != "" {
				if err := s.engine.ValidateGenerator(column.Industry, column.Generator); err != nil {
					return fmt.Errorf("invalid generator '%s' for industry '%s' in column %s of table %s: %w", column.Generator, column.Industry, column.Name, table.Name, err)
				}
			}
		}

		if err := s.db.CreateTable(ctx, table.Name, &table, schema.Relations); err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.Name, err)
		}

		// Create indexes
		for _, index := range table.Indexes {
			if err := s.db.CreateIndex(ctx, table.Name, index); err != nil {
				return fmt.Errorf("failed to create index %s on table %s: %w", index.Name, table.Name, err)
			}
		}
	}

	// Create constraints
	for _, constraint := range schema.Constraints {
		// Find the table that has the foreign key columns
		for _, table := range schema.Tables {
			hasColumns := true
			for _, col := range constraint.Columns {
				found := false
				for _, tableCol := range table.Columns {
					if tableCol.Name == col {
						found = true
						break
					}
				}
				if !found {
					hasColumns = false
					break
				}
			}
			if hasColumns {
				if err := s.db.CreateConstraint(ctx, table.Name, constraint); err != nil {
					return fmt.Errorf("failed to create constraint on table %s: %w", table.Name, err)
				}
				break
			}
		}
	}

	// Generate and insert data with proper foreign key relationships
	if err := s.generateAndInsertDataWithForeignKeys(ctx, schema); err != nil {
		return fmt.Errorf("failed to generate and insert data: %w", err)
	}

	return nil
}

// generateTableData generates data for a table based on its column specifications
func (s *Seeder) generateTableData(table *types.Table) ([]map[string]interface{}, error) {
	// Convert table columns to ColumnSpecs for the engine
	columnSpecs := make([]generators.ColumnSpec, len(table.Columns))
	for i, col := range table.Columns {
		columnSpecs[i] = generators.ColumnSpec{
			Name:      col.Name,
			Industry:  col.Industry,
			Generator: col.Generator,
			Params:    col.Params,
		}
	}
	
	// Generate the specified number of rows
	data := make([]map[string]interface{}, table.Count)
	for i := 0; i < table.Count; i++ {
		row, err := s.engine.GenerateRow(columnSpecs, make(map[string]interface{}))
		if err != nil {
			return nil, fmt.Errorf("failed to generate row %d: %w", i+1, err)
		}
		data[i] = row
	}
	
	return data, nil
}

// generateAndInsertDataWithForeignKeys generates and inserts data with proper foreign key relationships
func (s *Seeder) generateAndInsertDataWithForeignKeys(ctx context.Context, schema *types.Schema) error {
	// Store generated primary keys for foreign key references
	primaryKeys := make(map[string][]interface{}) // table_name -> array of primary key values
	
	// Determine table dependency order
	tableOrder, err := s.getTableDependencyOrder(schema)
	if err != nil {
		return fmt.Errorf("failed to determine table dependency order: %w", err)
	}
	
	log.Printf("Processing tables in dependency order: %v", tableOrder)
	
	// Process tables in dependency order
	for _, tableName := range tableOrder {
		table := s.findTableByName(schema, tableName)
		if table == nil {
			return fmt.Errorf("table %s not found in schema", tableName)
		}
		
		if table.Count <= 0 && len(table.Data) == 0 {
			continue // Skip tables with no data to generate
		}
		
		log.Printf("Generating %d rows for table: %s", table.Count, table.Name)
		
		var data []map[string]interface{}
		
		if table.Count > 0 {
			// Generate data with foreign key awareness and balanced distribution
			data, err = s.generateTableDataWithForeignKeys(table, schema, primaryKeys)
			if err != nil {
				return fmt.Errorf("failed to generate data for table %s: %w", table.Name, err)
			}
		} else {
			// Use pre-defined data
			data = table.Data
		}
		
		// Insert data
		log.Printf("Inserting %d rows into table '%s'", len(data), table.Name)
		if err := s.db.InsertData(ctx, table.Name, data); err != nil {
			return fmt.Errorf("failed to insert data into table %s: %w", table.Name, err)
		}
		
		// Store primary keys for foreign key references
		primaryKeyColumn := s.getPrimaryKeyColumn(table)
		if primaryKeyColumn != "" {
			keys := make([]interface{}, len(data))
			for i, row := range data {
				keys[i] = row[primaryKeyColumn]
			}
			primaryKeys[table.Name] = keys
			log.Printf("Stored %d primary keys for table %s", len(keys), table.Name)
		}
		
		log.Printf("Seeded table %s with %d rows", table.Name, len(data))
	}
	
	return nil
}

// getTableDependencyOrder returns tables in dependency order (parent tables first)
func (s *Seeder) getTableDependencyOrder(schema *types.Schema) ([]string, error) {
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
	
	// Topological sort using Kahn's algorithm
	var result []string
	inDegree := make(map[string]int)
	
	// Calculate in-degrees
	for table := range allTables {
		inDegree[table] = len(dependencies[table]) // Each table's in-degree is the number of dependencies it has
	}
	
	// Find tables with no dependencies (in-degree 0)
	queue := []string{}
	for table, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, table)
		}
	}
	
	// Process queue
	for len(queue) > 0 {
		// Remove first element
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)
		
		// Reduce in-degree for tables that depend on the current table
		for table, deps := range dependencies {
			for _, dep := range deps {
				if dep == current {
					inDegree[table]--
					if inDegree[table] == 0 {
						queue = append(queue, table)
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

// generateTableDataWithForeignKeys generates data for a table with foreign key awareness and balanced distribution
func (s *Seeder) generateTableDataWithForeignKeys(table *types.Table, schema *types.Schema, primaryKeys map[string][]interface{}) ([]map[string]interface{}, error) {
	// Convert table columns to ColumnSpecs for the engine
	columnSpecs := make([]generators.ColumnSpec, len(table.Columns))
	for i, col := range table.Columns {
		columnSpecs[i] = generators.ColumnSpec{
			Name:      col.Name,
			Industry:  col.Industry,
			Generator: col.Generator,
			Params:    col.Params,
		}
	}
	
	// Check if this table has foreign key relationships for balanced distribution
	hasForeignKeys := false
	var foreignKeyColumns []string
	var referencedTables []string
	
	for _, col := range table.Columns {
		if referencedTable, _, isForeignKey := config.GetForeignKeyInfo(schema, table.Name, col.Name); isForeignKey {
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
	
	var data []map[string]interface{}
	
	if hasForeignKeys && len(referencedTables) > 0 {
		// Use balanced distribution for tables with foreign keys
		log.Printf("Using balanced distribution for table %s (foreign keys found: %v)", table.Name, foreignKeyColumns)
		data = s.generateBalancedDistribution(columnSpecs, table, schema, primaryKeys, referencedTables)
	} else {
		// Use standard generation for parent tables (no foreign keys)
		log.Printf("Using standard generation for table %s (no foreign keys)", table.Name)
		data = make([]map[string]interface{}, table.Count)
		for i := 0; i < table.Count; i++ {
			row, err := s.engine.GenerateRow(columnSpecs, make(map[string]interface{}))
			if err != nil {
				return nil, fmt.Errorf("failed to generate row %d: %w", i+1, err)
			}
			data[i] = row
		}
	}
	
	return data, nil
}

// generateBalancedDistribution creates data with balanced foreign key distribution ensuring minimum records per referenced entity
func (s *Seeder) generateBalancedDistribution(columnSpecs []generators.ColumnSpec, table *types.Table, schema *types.Schema, primaryKeys map[string][]interface{}, referencedTables []string) []map[string]interface{} {
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
			row, _ := s.engine.GenerateRow(columnSpecs, make(map[string]interface{}))
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
		return s.generateEvenDistribution(columnSpecs, table, schema, primaryKeys, totalRecordsToGenerate, availablePrimaryKeys)
	}
	
	log.Printf("INFO: Balanced distribution for table %s: %d guaranteed + %d additional records across %d entities", 
		table.Name, guaranteedRecords, remainingRecords, numPrimaryKeys)
	
	recordIndex := 0
	
	// Phase 1: Guarantee minimum records per primary key
	for _, primaryKey := range availablePrimaryKeys {
		for i := 0; i < minRecordsPerReference && recordIndex < totalRecordsToGenerate; i++ {
			row, err := s.engine.GenerateRow(columnSpecs, make(map[string]interface{}))
			if err != nil {
				log.Printf("ERROR: Failed to generate row: %v", err)
				continue
			}
			
			// Set foreign key values
			s.setForeignKeyValues(row, table, schema, primaryKeys, primaryKey, primaryReferencedTable)
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
		selectedKey := s.selectWeightedRandomKey(availablePrimaryKeys, usageCounts)
		
		row, err := s.engine.GenerateRow(columnSpecs, make(map[string]interface{}))
		if err != nil {
			log.Printf("ERROR: Failed to generate row: %v", err)
			recordIndex++
			continue
		}
		
		// Set foreign key values
		s.setForeignKeyValues(row, table, schema, primaryKeys, selectedKey, primaryReferencedTable)
		data[recordIndex] = row
		usageCounts[selectedKey]++
		recordIndex++
	}
	
	// Log distribution summary
	s.logDistributionSummary(table.Name, usageCounts)
	
	return data
}

// generateEvenDistribution distributes records as evenly as possible across available primary keys
func (s *Seeder) generateEvenDistribution(columnSpecs []generators.ColumnSpec, table *types.Table, schema *types.Schema, primaryKeys map[string][]interface{}, totalRecords int, availablePrimaryKeys []interface{}) []map[string]interface{} {
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
	
	for i, primaryKey := range availablePrimaryKeys {
		numRecordsForThisKey := recordsPerKey
		if i < extraRecords {
			numRecordsForThisKey++ // Distribute extra records to first few keys
		}
		
		for j := 0; j < numRecordsForThisKey && recordIndex < totalRecords; j++ {
			row, err := s.engine.GenerateRow(columnSpecs, make(map[string]interface{}))
			if err != nil {
				log.Printf("ERROR: Failed to generate row: %v", err)
				recordIndex++
				continue
			}
			
			// Set foreign key values
			s.setForeignKeyValues(row, table, schema, primaryKeys, primaryKey, primaryReferencedTable)
			data[recordIndex] = row
			recordIndex++
		}
	}
	
	return data
}

// setForeignKeyValues sets foreign key values in a row based on the primary key and available references
func (s *Seeder) setForeignKeyValues(row map[string]interface{}, table *types.Table, schema *types.Schema, primaryKeys map[string][]interface{}, primaryKey interface{}, primaryReferencedTable string) {
	for _, col := range table.Columns {
		if referencedTable, referencedColumn, isForeignKey := config.GetForeignKeyInfo(schema, table.Name, col.Name); isForeignKey {
			if referencedTable == primaryReferencedTable {
				// Use the provided primary key for the primary reference
				row[col.Name] = primaryKey
				log.Printf("Set primary foreign key %s.%s = %v (referencing %s.%s)", table.Name, col.Name, primaryKey, referencedTable, referencedColumn)
			} else {
				// For secondary foreign keys, use random selection from available keys
				if availableKeys, exists := primaryKeys[referencedTable]; exists && len(availableKeys) > 0 {
					randomKey := availableKeys[rand.Intn(len(availableKeys))]
					row[col.Name] = randomKey
					log.Printf("Set secondary foreign key %s.%s = %v (referencing %s.%s)", table.Name, col.Name, randomKey, referencedTable, referencedColumn)
				}
			}
		}
	}
}

// selectWeightedRandomKey selects a key with preference for less used ones
func (s *Seeder) selectWeightedRandomKey(availableKeys []interface{}, usageCounts map[interface{}]int) interface{} {
	// Find minimum usage count
	minUsage := int(^uint(0) >> 1) // Max int
	for _, key := range availableKeys {
		if count := usageCounts[key]; count < minUsage {
			minUsage = count
		}
	}
	
	// Collect keys with minimum usage
	candidateKeys := make([]interface{}, 0)
	for _, key := range availableKeys {
		if usageCounts[key] == minUsage {
			candidateKeys = append(candidateKeys, key)
		}
	}
	
	// Randomly select from candidates
	if len(candidateKeys) > 0 {
		return candidateKeys[rand.Intn(len(candidateKeys))]
	}
	
	// Fallback to random selection
	return availableKeys[rand.Intn(len(availableKeys))]
}

// logDistributionSummary logs how records were distributed across primary keys
func (s *Seeder) logDistributionSummary(tableName string, usageCounts map[interface{}]int) {
	log.Printf("=== Distribution Summary for %s ===", tableName)
	for key, count := range usageCounts {
		log.Printf("  %v: %d records", key, count)
	}
	log.Printf("=== End Distribution Summary ===")
}

// findTableByName finds a table by name in the schema
func (s *Seeder) findTableByName(schema *types.Schema, tableName string) *types.Table {
	for i := range schema.Tables {
		if schema.Tables[i].Name == tableName {
			return &schema.Tables[i]
		}
	}
	return nil
}

// getPrimaryKeyColumn returns the name of the primary key column for a table
func (s *Seeder) getPrimaryKeyColumn(table *types.Table) string {
	for _, col := range table.Columns {
		if col.IsPrimary {
			return col.Name
		}
	}
	return ""
}
