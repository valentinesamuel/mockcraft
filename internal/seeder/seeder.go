package seeder

import (
	"context"
	"fmt"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
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
		if err := s.db.CreateTable(ctx, table.Name, table.Columns); err != nil {
			return fmt.Errorf("failed to create table %s: %w", table.Name, err)
		}

		// Create indexes
		for _, index := range table.Indexes {
			if err := s.db.CreateIndex(ctx, table.Name, index); err != nil {
				return fmt.Errorf("failed to create index %s on table %s: %w", index.Name, table.Name, err)
			}
		}

		// Insert data
		if len(table.Data) > 0 {
			if err := s.db.InsertData(ctx, table.Name, table.Data); err != nil {
				return fmt.Errorf("failed to insert data into table %s: %w", table.Name, err)
			}
		}
	}

	return nil
}
