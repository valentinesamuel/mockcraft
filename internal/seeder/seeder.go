package seeder

import (
	"context"
	"fmt"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"github.com/valentinesamuel/mockcraft/internal/generators"
	"github.com/valentinesamuel/mockcraft/internal/registry"
)

// Seeder represents a database seeder
type Seeder struct {
	db types.Database
}

// New creates a new database seeder
func New(db types.Database) *Seeder {
	return &Seeder{
		db: db,
	}
}

// Seed seeds the database with mock data
func (s *Seeder) Seed(ctx context.Context, schema *types.Schema) error {
	// Create tables
	for _, table := range schema.Tables {
		// Validate industry and generator for each column
		for _, column := range table.Columns {
			if column.Industry != "" {
				if _, err := registry.CreateGenerator(column.Industry); err != nil {
					return fmt.Errorf("invalid industry '%s' for column %s in table %s: %w", column.Industry, column.Name, table.Name, err)
				}
			}
			if column.Generator != "" {
				if _, err := generators.Get(column.Generator); err != nil {
					return fmt.Errorf("invalid generator '%s' for column %s in table %s: %w", column.Generator, column.Name, table.Name, err)
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

	// Insert data
	for _, table := range schema.Tables {
		if len(table.Data) > 0 {
			if err := s.db.InsertData(ctx, table.Name, table.Data); err != nil {
				return fmt.Errorf("failed to insert data into table %s: %w", table.Name, err)
			}
		}
	}

	return nil
}
