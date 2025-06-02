package database

import (
	"context"

	"github.com/valentinesamuel/mockcraft/internal/config"
)

// Seeder defines the interface for database seeding operations
type Seeder interface {
	// Connect establishes a connection to the database
	Connect(ctx context.Context) error
	// Close closes the database connection
	Close() error
	// SeedTable seeds a single table with generated data
	SeedTable(ctx context.Context, table *config.Table, data [][]interface{}) error
	// BeginTransaction starts a new transaction
	BeginTransaction(ctx context.Context) (Transaction, error)
}

// Transaction represents a database transaction
type Transaction interface {
	// Commit commits the transaction
	Commit() error
	// Rollback rolls back the transaction
	Rollback() error
}

// Factory creates database seeders based on connection string
type Factory interface {
	// NewSeeder creates a new seeder for the given connection string
	NewSeeder(dsn string) (Seeder, error)
}
