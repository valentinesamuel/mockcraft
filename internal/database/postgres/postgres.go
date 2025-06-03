package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

// PostgresDatabase implements the Database interface for PostgreSQL
type PostgresDatabase struct {
	config types.Config
	conn   *pgx.Conn
}

// NewPostgresDatabase creates a new PostgreSQL database connection
func NewPostgresDatabase(config types.Config) (*PostgresDatabase, error) {
	return &PostgresDatabase{
		config: config,
	}, nil
}

// Connect connects to the database
func (db *PostgresDatabase) Connect(ctx context.Context) error {
	// Build connection string
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		db.config.Username,
		db.config.Password,
		db.config.Host,
		db.config.Port,
		db.config.Database,
		db.config.SSLMode,
	)

	// Connect to database
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	db.conn = conn
	return nil
}

// Close closes the database connection
func (db *PostgresDatabase) Close() error {
	if db.conn != nil {
		return db.conn.Close(context.Background())
	}
	return nil
}

// CreateTable creates a table in the database
func (db *PostgresDatabase) CreateTable(ctx context.Context, tableName string, columns []types.Column) error {
	// Build CREATE TABLE statement
	var columnDefs []string
	for _, col := range columns {
		def := fmt.Sprintf("%s %s", col.Name, col.Type)
		if col.IsPrimary {
			def += " PRIMARY KEY"
		}
		if !col.IsNullable {
			def += " NOT NULL"
		}
		if col.IsUnique {
			def += " UNIQUE"
		}
		if col.Default != nil {
			def += fmt.Sprintf(" DEFAULT %v", col.Default)
		}
		columnDefs = append(columnDefs, def)
	}

	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s (%s)",
		tableName,
		strings.Join(columnDefs, ", "),
	)

	// Execute query
	if _, err := db.conn.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// CreateIndex creates an index on a table
func (db *PostgresDatabase) CreateIndex(ctx context.Context, tableName string, index types.Index) error {
	// Build CREATE INDEX statement
	query := fmt.Sprintf(
		"CREATE %s INDEX IF NOT EXISTS %s ON %s (%s)",
		index.Type,
		index.Name,
		tableName,
		strings.Join(index.Columns, ", "),
	)

	// Execute query
	if _, err := db.conn.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// InsertData inserts data into a table
func (db *PostgresDatabase) InsertData(ctx context.Context, tableName string, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Get column names from first row
	columns := make([]string, 0, len(data[0]))
	for k := range data[0] {
		columns = append(columns, k)
	}

	// Build INSERT statement
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(make([]string, len(columns)), ", "),
	)

	// Execute query for each row
	for _, row := range data {
		values := make([]interface{}, len(columns))
		for i, col := range columns {
			values[i] = row[col]
		}

		if _, err := db.conn.Exec(ctx, query, values...); err != nil {
			return fmt.Errorf("failed to insert data: %w", err)
		}
	}

	return nil
}

// BeginTransaction begins a new transaction
func (db *PostgresDatabase) BeginTransaction(ctx context.Context) (types.Transaction, error) {
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &PostgresTransaction{tx: tx}, nil
}

// GetDriver returns the database driver name
func (db *PostgresDatabase) GetDriver() string {
	return "postgres"
}

// PostgresTransaction implements the Transaction interface for PostgreSQL
type PostgresTransaction struct {
	tx pgx.Tx
}

// Commit commits the transaction
func (tx *PostgresTransaction) Commit() error {
	return tx.tx.Commit(context.Background())
}

// Rollback rolls back the transaction
func (tx *PostgresTransaction) Rollback() error {
	return tx.tx.Rollback(context.Background())
}
