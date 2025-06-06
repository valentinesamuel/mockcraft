package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

// PostgresDatabase implements the Database interface for PostgreSQL
type PostgresDatabase struct {
	config *types.Config
	conn   *pgx.Conn
}

// NewPostgresDatabase creates a new PostgreSQL database connection
func NewPostgresDatabase(config *types.Config) (*PostgresDatabase, error) {
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
func (db *PostgresDatabase) CreateTable(ctx context.Context, tableName string, table *types.Table, relations []types.Relationship) error {
	log.Printf("Creating table '%s'", tableName)

	var columnDefs []string
	for _, col := range table.Columns {
		def := fmt.Sprintf("%s %s", col.Name, db.getPostgresType(col.Type))
		if col.IsPrimary {
			def += " PRIMARY KEY"
		}
		if col.IsUnique {
			def += " UNIQUE"
		}
		if !col.IsNullable {
			def += " NOT NULL"
		}
		columnDefs = append(columnDefs, def)
	}

	// Add foreign key constraints
	var foreignKeyDefs []string
	for _, rel := range relations {
		// If this table is the 'to' table in a relationship, add a foreign key constraint
		if rel.ToTable == tableName {
			// CONSTRAINT constraint_name FOREIGN KEY (from_column) REFERENCES to_table (to_column)
			fkDef := fmt.Sprintf("CONSTRAINT fk_%s_%s_%s FOREIGN KEY (%s) REFERENCES %s (%s)",
				tableName,
				rel.ToColumn,
				rel.FromTable,
				rel.ToColumn,
				rel.FromTable,
				rel.FromColumn,
			)
			// Add ON DELETE and ON UPDATE clauses (assuming CASCADE for simplicity)
			fkDef += " ON DELETE CASCADE ON UPDATE CASCADE"
			foreignKeyDefs = append(foreignKeyDefs, fkDef)
		}
	}

	allDefs := append(columnDefs, foreignKeyDefs...)

	stmt := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", tableName, strings.Join(allDefs, ", "))

	log.Printf("Executing SQL: %s", stmt)

	_, err := db.conn.Exec(ctx, stmt)
	if err != nil {
		return fmt.Errorf("failed to create table '%s': %w", tableName, err)
	}

	log.Printf("Table '%s' created successfully.", tableName)

	return nil
}

// CreateIndex creates an index on a table
func (db *PostgresDatabase) CreateIndex(ctx context.Context, tableName string, index types.Index) error {
	log.Printf("Creating index '%s' on table '%s'", index.Name, tableName)

	indexType := ""
	if index.Type != "" {
		indexType = " USING " + index.Type
	}

	unique := ""
	if index.IsUnique {
		unique = "UNIQUE "
	}

	stmt := fmt.Sprintf("CREATE %sINDEX IF NOT EXISTS %s ON %s (%s)%s",
		unique,
		index.Name,
		tableName,
		strings.Join(index.Columns, ", "),
		indexType,
	)

	log.Printf("Executing SQL: %s", stmt)

	_, err := db.conn.Exec(ctx, stmt)
	if err != nil {
		return fmt.Errorf("failed to create index '%s': %w", index.Name, err)
	}

	log.Printf("Index '%s' created successfully.", index.Name)

	return nil
}

// CreateConstraint creates a constraint on a table (e.g., foreign key)
func (db *PostgresDatabase) CreateConstraint(ctx context.Context, tableName string, constraint types.Constraint) error {
	log.Printf("Creating constraint of type '%s' on table '%s' for columns %v", constraint.Type, tableName, constraint.Columns)

	log.Printf("Skipping constraint creation in Postgres. Should be handled via ALTER TABLE based on schema.Relations.")
	return nil
}

// InsertData inserts data into a table
func (db *PostgresDatabase) InsertData(ctx context.Context, tableName string, data []map[string]interface{}) error {
	log.Printf("Inserting %d rows into table '%s'", len(data), tableName)

	if len(data) == 0 {
		log.Printf("No data to insert into '%s'", tableName)
		return nil
	}

	cols := make([]string, 0, len(data[0]))
	for colName := range data[0] {
		cols = append(cols, colName)
	}

	var valuePlaceholders []string
	var values []interface{}

	for _, row := range data {
		var rowValues []string
		for _, colName := range cols {
			rowValues = append(rowValues, fmt.Sprintf("$%d", len(values)+1))
			values = append(values, row[colName])
		}
		valuePlaceholders = append(valuePlaceholders, fmt.Sprintf("(%s)", strings.Join(rowValues, ", ")))
	}

	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(cols, ", "),
		strings.Join(valuePlaceholders, ", "),
	)

	log.Printf("Executing SQL: %s", stmt)

	_, err := db.conn.Exec(ctx, stmt, values...)
	if err != nil {
		return fmt.Errorf("failed to insert data into '%s': %w", tableName, err)
	}

	log.Printf("%d rows inserted into '%s'.", len(data), tableName)

	return nil
}

// UpdateData updates existing data in a table (placeholder)
func (db *PostgresDatabase) UpdateData(ctx context.Context, tableName string, data []map[string]interface{}) error {
	log.Printf("Updating %d rows in table '%s' (placeholder implementation)", len(data), tableName)
	return fmt.Errorf("UpdateData not fully implemented for PostgreSQL")
}

// GetAllIDs retrieves all primary key IDs from a table
func (db *PostgresDatabase) GetAllIDs(ctx context.Context, tableName string) ([]string, error) {
	log.Printf("Getting all IDs from table '%s'", tableName)

	stmt := fmt.Sprintf("SELECT id FROM %s", tableName)

	rows, err := db.conn.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to get IDs from '%s': %w", tableName, err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan ID from row: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("Retrieved %d IDs from '%s'", len(ids), tableName)

	return ids, nil
}

// GetAllForeignKeys retrieves all values from a foreign key column
func (db *PostgresDatabase) GetAllForeignKeys(ctx context.Context, tableName string, columnName string) ([]string, error) {
	log.Printf("Getting all foreign key values from '%s'.'%s'", tableName, columnName)

	stmt := fmt.Sprintf("SELECT %s FROM %s", columnName, tableName)

	rows, err := db.conn.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to get foreign keys from '%s'.'%s': %w", tableName, columnName, err)
	}
	defer rows.Close()

	var fks []string
	for rows.Next() {
		var fk sql.NullString
		if err := rows.Scan(&fk); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key from row: %w", err)
		}
		if fk.Valid {
			fks = append(fks, fk.String)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("Retrieved %d foreign key values from '%s'.'%s'", len(fks), tableName, columnName)

	return fks, nil
}

// VerifyReferentialIntegrity checks if foreign key references are valid
func (db *PostgresDatabase) VerifyReferentialIntegrity(ctx context.Context, fromTable, fromColumn, toTable, toColumn string) error {
	log.Printf("Verifying referential integrity: %s.%s -> %s.%s", fromTable, fromColumn, toTable, toColumn)

	stmt := fmt.Sprintf(`
SELECT COUNT(*)
FROM %s t
LEFT JOIN %s f ON t.%s = f.%s
WHERE t.%s IS NOT NULL AND f.%s IS NULL`, // Count rows in child where FK is not null but parent does not exist
		fromTable, toTable, fromColumn, toColumn, fromColumn, toColumn,
	)

	log.Printf("Executing integrity check SQL: %s", stmt)

	var invalidCount int
	err := db.conn.QueryRow(ctx, stmt).Scan(&invalidCount)
	if err != nil {
		return fmt.Errorf("failed to execute integrity check query: %w", err)
	}

	if invalidCount > 0 {
		return fmt.Errorf("found %d invalid foreign key references in %s.%s", invalidCount, fromTable, toColumn)
	}

	log.Printf("Referential integrity check passed for %s.%s -> %s.%s", fromTable, fromColumn, toTable, toColumn)

	return nil
}

// DropTable drops a table from the database
func (db *PostgresDatabase) DropTable(ctx context.Context, tableName string) error {
	log.Printf("Dropping table '%s'", tableName)

	stmt := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tableName)

	log.Printf("Executing SQL: %s", stmt)

	_, err := db.conn.Exec(ctx, stmt)
	if err != nil {
		return fmt.Errorf("failed to drop table '%s': %w", tableName, err)
	}

	log.Printf("Table '%s' dropped successfully.", tableName)

	return nil
}

// GetDriver returns the database driver name
func (db *PostgresDatabase) GetDriver() string {
	return "postgres"
}

// BeginTransaction begins a new transaction (placeholder)
func (db *PostgresDatabase) BeginTransaction(ctx context.Context) (types.Transaction, error) {
	log.Println("Beginning transaction (placeholder)")
	return nil, fmt.Errorf("BeginTransaction not implemented for PostgreSQL")
}

// getPostgresType maps schema types to PostgreSQL types
func (db *PostgresDatabase) getPostgresType(schemaType string) string {
	switch strings.ToLower(schemaType) {
	case "string", "text", "uuid":
		return "TEXT"
	case "integer", "int":
		return "INT"
	case "number":
		return "NUMERIC"
	case "decimal":
		return "DECIMAL"
	case "float":
		return "FLOAT"
	case "boolean":
		return "BOOLEAN"
	case "timestamp", "datetime":
		return "TIMESTAMP WITH TIME ZONE"
	case "date":
		return "DATE"
	default:
		log.Printf("Warning: Unknown schema type '%s', mapping to TEXT", schemaType)
		return "TEXT"
	}
}
