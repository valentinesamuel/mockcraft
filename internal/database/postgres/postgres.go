package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/valentinesamuel/mockcraft/internal/config"
	"github.com/valentinesamuel/mockcraft/internal/database"
)

type postgresSeeder struct {
	db  *sql.DB
	dsn string
}

type postgresTransaction struct {
	tx *sql.Tx
}

// NewSeeder creates a new PostgreSQL seeder
func NewSeeder(dsn string) (database.Seeder, error) {
	return &postgresSeeder{dsn: dsn}, nil
}

func (p *postgresSeeder) Connect(ctx context.Context) error {
	db, err := sql.Open("postgres", p.dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	p.db = db
	return nil
}

func (p *postgresSeeder) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *postgresSeeder) BeginTransaction(ctx context.Context) (database.Transaction, error) {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}
	return &postgresTransaction{tx: tx}, nil
}

func (p *postgresSeeder) SeedTable(ctx context.Context, table *config.Table, data [][]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Build column names
	columns := make([]string, len(table.Columns))
	for i, col := range table.Columns {
		columns[i] = col.Name
	}

	// Build placeholders for the query
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	// Build the INSERT query
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		pq.QuoteIdentifier(table.Name),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// Begin transaction
	tx, err := p.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Prepare the statement
	stmt, err := p.db.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	// Insert each row
	for _, row := range data {
		_, err := stmt.ExecContext(ctx, row...)
		if err != nil {
			return fmt.Errorf("failed to insert row: %v", err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (t *postgresTransaction) Commit() error {
	return t.tx.Commit()
}

func (t *postgresTransaction) Rollback() error {
	return t.tx.Rollback()
}
