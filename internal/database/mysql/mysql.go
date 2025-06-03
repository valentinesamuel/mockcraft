package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

// MySQLDatabase represents a MySQL database connection
type MySQLDatabase struct {
	config types.Config
	db     *sql.DB
}

// NewMySQLDatabase creates a new MySQL database connection
func NewMySQLDatabase(config types.Config) (*MySQLDatabase, error) {
	return &MySQLDatabase{
		config: config,
	}, nil
}

// Connect establishes a connection to the MySQL database
func (m *MySQLDatabase) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.config.Username,
		m.config.Password,
		m.config.Host,
		m.config.Port,
		m.config.Database,
	)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(m.config.MaxOpenConns)
	conn.SetMaxIdleConns(m.config.MaxIdleConns)
	conn.SetConnMaxLifetime(m.config.ConnMaxLifetime)
	conn.SetConnMaxIdleTime(m.config.ConnMaxIdleTime)

	// Test connection
	if err := conn.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	m.db = conn
	return nil
}

// Close closes the MySQL database connection
func (m *MySQLDatabase) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// CreateTable creates a table in the MySQL database
func (m *MySQLDatabase) CreateTable(ctx context.Context, tableName string, columns []types.Column) error {
	var columnDefs []string
	for _, col := range columns {
		def := fmt.Sprintf("`%s` %s", col.Name, m.getMySQLType(col.Type))
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

	_, err := m.db.ExecContext(ctx, query)
	return err
}

// CreateIndex creates an index on a MySQL table
func (m *MySQLDatabase) CreateIndex(ctx context.Context, tableName string, index types.Index) error {
	if len(index.Columns) == 0 {
		return fmt.Errorf("no columns specified for index")
	}

	// Build index definition
	indexType := "INDEX"
	if index.IsUnique {
		indexType = "UNIQUE INDEX"
	}

	// Add index type if specified
	if index.Type != "" {
		indexType = fmt.Sprintf("%s USING %s", indexType, strings.ToUpper(index.Type))
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

	_, err := m.db.ExecContext(ctx, query)
	return err
}

// InsertData inserts data into a MySQL table
func (m *MySQLDatabase) InsertData(ctx context.Context, tableName string, data []map[string]interface{}) error {
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
	stmt, err := m.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Begin transaction
	tx, err := m.db.BeginTx(ctx, nil)
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

// BeginTransaction starts a new MySQL transaction
func (m *MySQLDatabase) BeginTransaction(ctx context.Context) (types.Transaction, error) {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &MySQLTransaction{tx: tx}, nil
}

// GetDriver returns the MySQL driver name
func (m *MySQLDatabase) GetDriver() string {
	return "mysql"
}

// getMySQLType converts a generic type to MySQL type
func (m *MySQLDatabase) getMySQLType(typ string) string {
	switch strings.ToLower(typ) {
	case "uuid":
		return "CHAR(36)"
	case "string":
		return "VARCHAR(255)"
	case "text":
		return "TEXT"
	case "integer":
		return "INT"
	case "decimal":
		return "DECIMAL(10,2)"
	case "boolean":
		return "BOOLEAN"
	case "timestamp":
		return "TIMESTAMP"
	default:
		return typ
	}
}

// MySQLTransaction represents a MySQL transaction
type MySQLTransaction struct {
	tx *sql.Tx
}

// Commit commits the MySQL transaction
func (t *MySQLTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the MySQL transaction
func (t *MySQLTransaction) Rollback() error {
	return t.tx.Rollback()
}
