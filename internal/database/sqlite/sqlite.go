package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

// SQLite represents a SQLite database connection
type SQLite struct {
	config *types.Config
	db     *sql.DB
}

// New creates a new SQLite database connection
func NewSQLiteDatabase(config *types.Config) (*SQLite, error) {
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
func (s *SQLite) CreateTable(ctx context.Context, tableName string, table *types.Table, relations []types.Relationship) error {
	log.Printf("Creating table '%s'", tableName)

	// Start building the CREATE TABLE statement
	var columns []string
	for _, col := range table.Columns {
		// Map the schema type to SQLite type
		sqliteType := s.getSQLiteType(col.Type)

		// Build column definition
		columnDef := fmt.Sprintf("`%s` %s", col.Name, sqliteType)

		// Add NOT NULL if specified
		if !col.IsNullable {
			columnDef += " NOT NULL"
		}

		// Add PRIMARY KEY if specified
		if col.IsPrimary {
			columnDef += " PRIMARY KEY"
		}

		// Add UNIQUE if specified
		if col.IsUnique {
			columnDef += " UNIQUE"
		}

		// Add DEFAULT if specified
		if col.Default != nil {
			columnDef += fmt.Sprintf(" DEFAULT %v", col.Default)
		}

		columns = append(columns, columnDef)
	}

	// Add foreign key constraints
	for _, rel := range relations {
		if rel.ToTable == tableName {
			// This is a foreign key in the current table
			fkDef := fmt.Sprintf("FOREIGN KEY (`%s`) REFERENCES `%s` (`%s`) ON DELETE CASCADE ON UPDATE CASCADE",
				rel.ToColumn, rel.FromTable, rel.FromColumn)
			columns = append(columns, fkDef)
		}
	}

	// Create the table
	createStmt := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (%s)",
		tableName, strings.Join(columns, ", "))

	log.Printf("Executing SQL: %s", createStmt)
	_, err := s.db.ExecContext(ctx, createStmt)
	if err != nil {
		return fmt.Errorf("failed to create table '%s': %w", tableName, err)
	}

	log.Printf("Table '%s' created successfully.", tableName)
	return nil
}

// CreateConstraint creates a constraint on a table
func (s *SQLite) CreateConstraint(ctx context.Context, tableName string, constraint types.Constraint) error {
	if constraint.Type == "foreign_key" {
		// SQLite doesn't support adding constraints after table creation
		// We need to recreate the table with the constraint
		// First, get the current table structure
		rows, err := s.db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(`%s`)", tableName))
		if err != nil {
			return fmt.Errorf("failed to get table info: %w", err)
		}
		defer rows.Close()

		var columns []string
		for rows.Next() {
			var cid, notnull, pk int
			var name, type_, dflt_value string
			if err := rows.Scan(&cid, &name, &type_, &notnull, &dflt_value, &pk); err != nil {
				return fmt.Errorf("failed to scan table info: %w", err)
			}
			columns = append(columns, fmt.Sprintf("`%s` %s", name, type_))
		}

		// Add the foreign key constraint
		columns = append(columns, fmt.Sprintf("FOREIGN KEY (`%s`) %s",
			strings.Join(constraint.Columns, "`, `"),
			constraint.Condition,
		))

		// Create new table with constraint
		query := fmt.Sprintf("CREATE TABLE `%s_new` (%s)",
			tableName,
			strings.Join(columns, ", "),
		)

		if _, err := s.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to create new table with constraint: %w", err)
		}

		// Copy data
		if _, err := s.db.ExecContext(ctx, fmt.Sprintf("INSERT INTO `%s_new` SELECT * FROM `%s`", tableName, tableName)); err != nil {
			return fmt.Errorf("failed to copy data: %w", err)
		}

		// Drop old table
		if _, err := s.db.ExecContext(ctx, fmt.Sprintf("DROP TABLE `%s`", tableName)); err != nil {
			return fmt.Errorf("failed to drop old table: %w", err)
		}

		// Rename new table
		if _, err := s.db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE `%s_new` RENAME TO `%s`", tableName, tableName)); err != nil {
			return fmt.Errorf("failed to rename new table: %w", err)
		}
	}
	return nil
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
	log.Printf("Inserting %d rows into table '%s'", len(data), tableName)

	if len(data) == 0 {
		log.Printf("No data to insert into '%s'", tableName)
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

	// Begin transaction for bulk insert performance
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Execute the statement for each row within the transaction
	for _, row := range data {
		values := make([]interface{}, len(columns))
		for i, col := range columns {
			values[i] = row[col]
		}

		_, err := tx.Stmt(stmt).ExecContext(ctx, values...)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert row into '%s': %w", tableName, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for '%s': %w", tableName, err)
	}

	log.Printf("%d rows inserted into '%s'.", len(data), tableName)

	return nil
}

// UpdateData updates existing data in a table (placeholder for SQLite)
func (s *SQLite) UpdateData(ctx context.Context, tableName string, data []map[string]interface{}) error {
	log.Printf("Updating %d rows in table '%s' (placeholder implementation)", len(data), tableName)
	// SQLite updates can be done with UPDATE statements, typically by primary key.
	// This is a placeholder and needs a proper implementation.
	return fmt.Errorf("UpdateData not fully implemented for SQLite")
}

// GetAllIDs retrieves all primary key IDs from a table
func (s *SQLite) GetAllIDs(ctx context.Context, tableName string) ([]string, error) {
	// Query to get all IDs from the table using _id
	query := fmt.Sprintf("SELECT _id FROM %s", tableName)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get IDs from '%s': %w", tableName, err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan ID: %w", err)
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over IDs: %w", err)
	}

	return ids, nil
}

// GetAllForeignKeys retrieves all values from a foreign key column
func (s *SQLite) GetAllForeignKeys(ctx context.Context, tableName string, columnName string) ([]string, error) {
	log.Printf("Getting all foreign key values from '%s'.'%s'", tableName, columnName)

	stmt := fmt.Sprintf("SELECT `%s` FROM `%s`", columnName, tableName)

	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to get foreign keys from '%s'.'%s': %w", tableName, columnName, err)
	}
	defer rows.Close()

	var fks []string
	for rows.Next() {
		// Handle potential NULL values in foreign keys
		var nullableFK sql.NullString
		if err := rows.Scan(&nullableFK); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key from row: %w", err)
		}
		if nullableFK.Valid {
			fks = append(fks, nullableFK.String)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	log.Printf("Retrieved %d foreign key values from '%s'.'%s'", len(fks), tableName, columnName)

	return fks, nil
}

// VerifyReferentialIntegrity checks if foreign key references are valid
func (s *SQLite) VerifyReferentialIntegrity(ctx context.Context, fromTable, fromColumn, toTable, toColumn string) error {
	log.Printf("Verifying referential integrity: %s.%s -> %s.%s", fromTable, fromColumn, toTable, toColumn)

	// Check for orphaned records in the 'toTable' (child) that reference non-existent records in 'fromTable' (parent)
	stmt := fmt.Sprintf(`
SELECT COUNT(*)
FROM %s t
LEFT JOIN %s f ON t.%s = f.%s
WHERE t.%s IS NOT NULL AND f.%s IS NULL`, // Count rows in child where FK is not null but parent does not exist
		"`"+toTable+"`", "`"+fromTable+"`", "`"+toColumn+"`", "`"+fromColumn+"`", "`"+toColumn+"`", "`"+fromColumn+"`",
	)

	log.Printf("Executing integrity check SQL: %s", stmt)

	var invalidCount int
	err := s.db.QueryRowContext(ctx, stmt).Scan(&invalidCount)
	if err != nil {
		return fmt.Errorf("failed to execute integrity check query: %w", err)
	}

	if invalidCount > 0 {
		return fmt.Errorf("found %d invalid foreign key references in %s.%s", invalidCount, toTable, toColumn)
	}

	log.Printf("Referential integrity check passed for %s.%s -> %s.%s", fromTable, fromColumn, toTable, toColumn)

	return nil
}

// DropTable drops a table from the database
func (s *SQLite) DropTable(ctx context.Context, tableName string) error {
	log.Printf("Dropping table '%s'", tableName)

	stmt := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)

	log.Printf("Executing SQL: %s", stmt)

	_, err := s.db.ExecContext(ctx, stmt)
	if err != nil {
		return fmt.Errorf("failed to drop table '%s': %w", tableName, err)
	}

	log.Printf("Table '%s' dropped successfully.", tableName)

	return nil
}

// GetDriver returns the database driver name
func (s *SQLite) GetDriver() string {
	return "sqlite"
}

// BeginTransaction begins a new transaction
func (s *SQLite) BeginTransaction(ctx context.Context) (types.Transaction, error) {
	log.Println("Beginning transaction")
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return &SQLiteTransaction{tx: tx}, nil
}

// getSQLiteType maps schema types to SQLite types
func (s *SQLite) getSQLiteType(schemaType string) string {
	switch strings.ToLower(schemaType) {
	// Text affinity types
	case "string", "text", "uuid", "char", "varchar", "clob":
		return "TEXT"
	case "json", "jsonb": // JSON stored as TEXT in SQLite
		return "TEXT"
	case "xml": // XML stored as TEXT
		return "TEXT"
	
	// Integer affinity types
	case "integer", "int", "number":
		return "INTEGER"
	case "tinyint", "smallint", "mediumint", "bigint":
		return "INTEGER"
	case "int2", "int8":
		return "INTEGER"
	case "unsigned big int":
		return "INTEGER"
	case "serial", "bigserial": // Auto-incrementing integers
		return "INTEGER"
	case "boolean":
		return "INTEGER" // SQLite uses 0 for false, 1 for true
	
	// Real affinity types  
	case "real", "double", "double precision", "float":
		return "REAL"
	case "decimal", "numeric": // Numeric values can be REAL or TEXT
		return "REAL"
	case "money": // Store as REAL for calculations
		return "REAL"
	
	// Date/Time types (stored as TEXT in ISO8601 format)
	case "date", "datetime", "timestamp":
		return "TEXT"
	case "time": 
		return "TEXT"
	case "interval": // Duration stored as TEXT
		return "TEXT"
	
	// Binary data
	case "blob", "bytea", "binary":
		return "BLOB"
	
	// Array types (stored as JSON TEXT in SQLite)
	case "array", "text[]", "integer[]", "boolean[]":
		return "TEXT" // Store as JSON array
	
	// Network types (stored as TEXT)
	case "inet", "cidr", "macaddr":
		return "TEXT"
	
	// Geometric types (stored as TEXT, could be JSON or custom format)
	case "point", "line", "circle", "polygon", "path", "box", "lseg":
		return "TEXT"
	
	// Range types (stored as TEXT)
	case "int4range", "int8range", "numrange", "tsrange", "tstzrange", "daterange":
		return "TEXT"
	
	// Full-text search (stored as TEXT)
	case "tsvector", "tsquery":
		return "TEXT"
	
	// Bit strings (stored as TEXT or BLOB)
	case "bit", "bit varying":
		return "TEXT"
	
	// Handle parameterized types (varchar(n), char(n), etc.)
	default:
		lowerType := strings.ToLower(schemaType)
		if strings.HasPrefix(lowerType, "varchar") ||
		   strings.HasPrefix(lowerType, "char") ||
		   strings.HasPrefix(lowerType, "text") {
			return "TEXT"
		}
		if strings.HasPrefix(lowerType, "decimal") ||
		   strings.HasPrefix(lowerType, "numeric") {
			return "REAL"
		}
		if strings.HasPrefix(lowerType, "int") ||
		   strings.HasPrefix(lowerType, "bigint") ||
		   strings.HasPrefix(lowerType, "smallint") {
			return "INTEGER"
		}
		if strings.HasSuffix(lowerType, "[]") {
			return "TEXT" // Arrays stored as JSON
		}
		
		log.Printf("Warning: Unknown schema type '%s', mapping to TEXT", schemaType)
		return "TEXT"
	}
}

type SQLiteTransaction struct {
	tx *sql.Tx
}

// Commit commits the transaction
func (t *SQLiteTransaction) Commit() error {
	return t.tx.Commit()
}

// Rollback rolls back the transaction
func (t *SQLiteTransaction) Rollback() error {
	return t.tx.Rollback()
}

// Backup creates a backup of the database
func (s *SQLite) Backup(ctx context.Context, backupPath string) error {
	log.Printf("Creating backup of database '%s' to '%s'", s.config.Database, backupPath)

	// Create the backup directory if it doesn't exist
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Open the source database file
	srcFile, err := os.Open(s.config.Database)
	if err != nil {
		return fmt.Errorf("failed to open source database file: %w", err)
	}
	defer srcFile.Close()

	// Create the backup file
	dstFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer dstFile.Close()

	// Copy the database file
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy database file: %w", err)
	}

	// Ensure all data is written to disk
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync backup file: %w", err)
	}

	log.Printf("Backup created successfully at '%s'", backupPath)
	return nil
}

// Restore restores the database from a backup file
func (s *SQLite) Restore(ctx context.Context, backupPath string) error {
	log.Printf("Restoring database from '%s' to '%s'", backupPath, s.config.Database)

	// Close the current database connection before replacing the file
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("failed to close database connection before restore: %w", err)
		}
		s.db = nil // Set to nil so Connect will open a new connection later
	}

	// Open the backup file
	srcFile, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file '%s': %w", backupPath, err)
	}
	defer srcFile.Close()

	// Create or open the target database file
	dstFile, err := os.Create(s.config.Database)
	if err != nil {
		return fmt.Errorf("failed to create database file '%s' for restore: %w", s.config.Database, err)
	}
	defer dstFile.Close()

	// Copy the content of the backup file to the database file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy backup data to database file: %w", err)
	}

	// Ensure data is synced to disk
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync database file after restore: %w", err)
	}

	log.Printf("Database restored successfully from '%s'", backupPath)

	// Re-establish connection after restore
	if err := s.Connect(ctx); err != nil {
		return fmt.Errorf("failed to reconnect to database after restore: %w", err)
	}

	return nil
}
