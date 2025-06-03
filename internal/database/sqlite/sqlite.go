package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

// SQLite represents a SQLite database connection
type SQLite struct {
	config types.Config
	db     *sql.DB
}

// New creates a new SQLite database connection
func NewSQLiteDatabase(config types.Config) (*SQLite, error) {
	return &SQLite{
		config: config,
	}, nil
}

// Connect establishes a connection to the SQLite database
func (s *SQLite) Connect(ctx context.Context) error {
	db, err := sql.Open("sqlite3", s.config.Database)
	if err != nil {
		return fmt.Errorf("failed to open SQLite connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(s.config.MaxOpenConns)
	db.SetMaxIdleConns(s.config.MaxIdleConns)
	db.SetConnMaxLifetime(s.config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(s.config.ConnMaxIdleTime)

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	s.db = db
	return nil
}

// Close closes the SQLite database connection
func (s *SQLite) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// CreateTable creates a table in the SQLite database
func (s *SQLite) CreateTable(ctx context.Context, tableName string, columns []types.Column) error {
	var columnDefs []string
	for _, col := range columns {
		def := fmt.Sprintf("`%s` %s", col.Name, s.getSQLiteType(col.Type))
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

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (%s)",
		tableName,
		strings.Join(columnDefs, ", "),
	)

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// CreateIndex creates an index on a SQLite table
func (s *SQLite) CreateIndex(ctx context.Context, tableName string, index types.Index) error {
	if len(index.Columns) == 0 {
		return fmt.Errorf("no columns specified for index")
	}

	// Build index definition
	indexType := "INDEX"
	if index.IsUnique {
		indexType = "UNIQUE INDEX"
	}

	// Build the CREATE INDEX query
	query := fmt.Sprintf("CREATE %s `%s` ON `%s` (`%s`)",
		indexType,
		index.Name,
		tableName,
		strings.Join(index.Columns, "`, `"),
	)

	// Add additional properties if specified
	if len(index.Properties) > 0 {
		var props []string
		for key, value := range index.Properties {
			props = append(props, fmt.Sprintf("%s = %v", key, value))
		}
		if len(props) > 0 {
			query += " " + strings.Join(props, ", ")
		}
	}

	_, err := s.db.ExecContext(ctx, query)
	return err
}

// InsertData inserts data into a SQLite table
func (s *SQLite) InsertData(ctx context.Context, tableName string, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Get column names from the first row
	var columns []string
	for col := range data[0] {
		columns = append(columns, col)
	}

	// Build the INSERT query
	placeholders := make([]string, len(columns))
	for i := range columns {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("INSERT INTO `%s` (`%s`) VALUES (%s)",
		tableName,
		strings.Join(columns, "`, `"),
		strings.Join(placeholders, ", "),
	)

	// Prepare the statement
	stmt, err := s.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Begin transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Insert each row
	for _, row := range data {
		args := make([]interface{}, len(columns))
		for i, col := range columns {
			args[i] = row[col]
		}

		_, err := tx.StmtContext(ctx, stmt).ExecContext(ctx, args...)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// BeginTransaction starts a new SQLite transaction
func (s *SQLite) BeginTransaction(ctx context.Context) (types.Transaction, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &SQLiteTransaction{tx: tx}, nil
}

// GetDriver returns the SQLite driver name
func (s *SQLite) GetDriver() string {
	return "sqlite3"
}

// getSQLiteType converts a generic type to SQLite type
func (s *SQLite) getSQLiteType(typ string) string {
	switch strings.ToLower(typ) {
	case "uuid":
		return "TEXT"
	case "string":
		return "TEXT"
	case "text":
		return "TEXT"
	case "integer":
		return "INTEGER"
	case "decimal":
		return "REAL"
	case "boolean":
		return "INTEGER"
	case "timestamp":
		return "DATETIME"
	default:
		return typ
	}
}

// SQLiteTransaction represents a SQLite transaction
type SQLiteTransaction struct {
	tx *sql.Tx
}

// Commit commits the SQLite transaction
func (t *SQLiteTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the SQLite transaction
func (t *SQLiteTransaction) Rollback() error {
	return t.tx.Rollback()
}
