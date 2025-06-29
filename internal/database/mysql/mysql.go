package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

// MySQLDatabase represents a MySQL database connection
type MySQLDatabase struct {
	config *types.Config
	db     *sql.DB
}

// NewMySQLDatabase creates a new MySQL database connection
func NewMySQLDatabase(config *types.Config) (*MySQLDatabase, error) {
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

// CreateTable creates a table in the database
func (db *MySQLDatabase) CreateTable(ctx context.Context, tableName string, table *types.Table, relations []types.Relationship) error {
	log.Printf("Creating table '%s'", tableName)

	var columnDefs []string
	for _, col := range table.Columns {
		def := fmt.Sprintf("`%s` %s", col.Name, db.getMySQLType(col.Type))
		if col.IsPrimary {
			def += " PRIMARY KEY"
			// Add AUTO_INCREMENT for integer primary keys if needed
			if col.Type == "integer" || col.Type == "int" {
				def += " AUTO_INCREMENT"
			}
		}
		if col.IsUnique {
			def += " UNIQUE"
		}
		if !col.IsNullable {
			def += " NOT NULL"
		}
		// Default values are handled by generators during data insertion, not typically in schema DDL for this tool's approach.
		columnDefs = append(columnDefs, def)
	}

	// Add foreign key constraints
	var foreignKeyDefs []string
	for _, rel := range relations {
		// If this table is the 'to' table in a relationship, add a foreign key constraint
		if rel.ToTable == tableName {
			// CONSTRAINT constraint_name FOREIGN KEY (from_column) REFERENCES to_table (to_column)
			fkDef := fmt.Sprintf("CONSTRAINT fk_%s_%s_%s FOREIGN KEY (`%s`) REFERENCES `%s` (`%s`)",
				tableName,
				rel.ToColumn,
				rel.FromTable,
				rel.ToColumn,
				rel.FromTable,
				rel.FromColumn,
			)
			// Add ON DELETE and ON UPDATE clauses if specified in the relationship (assuming a field like rel.OnDelete/rel.OnUpdate exists or adding CASCADE as default)
			// For simplicity, adding ON DELETE CASCADE and ON UPDATE CASCADE as a common pattern
			fkDef += " ON DELETE CASCADE ON UPDATE CASCADE"
			foreignKeyDefs = append(foreignKeyDefs, fkDef)
		}
	}

	allDefs := append(columnDefs, foreignKeyDefs...)

	stmt := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (%s)", tableName, strings.Join(allDefs, ", "))

	log.Printf("Executing SQL: %s", stmt)

	_, err := db.db.ExecContext(ctx, stmt)
	if err != nil {
		return fmt.Errorf("failed to create table '%s': %w", tableName, err)
	}

	log.Printf("Table '%s' created successfully.", tableName)

	return nil
}

// CreateConstraint creates a constraint on a table
func (m *MySQLDatabase) CreateConstraint(ctx context.Context, tableName string, constraint types.Constraint) error {
	if constraint.Type == "foreign_key" {
		query := fmt.Sprintf(
			"ALTER TABLE `%s` ADD CONSTRAINT fk_%s_%s FOREIGN KEY (`%s`) %s",
			tableName,
			tableName,
			strings.Join(constraint.Columns, "_"),
			strings.Join(constraint.Columns, "`, `"),
			constraint.Condition,
		)

		_, err := m.db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create foreign key constraint: %w", err)
		}
	}
	return nil
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

// DropTable drops a table in MySQL
func (m *MySQLDatabase) DropTable(ctx context.Context, tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := m.db.ExecContext(ctx, query)
	return err
}

// getMySQLType maps schema types to MySQL types
func (db *MySQLDatabase) getMySQLType(schemaType string) string {
	switch strings.ToLower(schemaType) {
	// String and text types
	case "string", "uuid":
		return "VARCHAR(255)"
	case "text":
		return "TEXT"
	case "char":
		return "CHAR(1)"
	case "tinytext":
		return "TINYTEXT"
	case "mediumtext":
		return "MEDIUMTEXT"
	case "longtext":
		return "LONGTEXT"
	
	// Integer types
	case "integer", "int", "number":
		return "INT"
	case "tinyint":
		return "TINYINT"
	case "smallint":
		return "SMALLINT"
	case "mediumint":
		return "MEDIUMINT"
	case "bigint":
		return "BIGINT"
	case "serial":
		return "INT AUTO_INCREMENT"
	case "bigserial":
		return "BIGINT AUTO_INCREMENT"
	
	// Decimal and floating point types
	case "decimal", "numeric":
		return "DECIMAL(10,2)"
	case "float":
		return "FLOAT"
	case "double":
		return "DOUBLE"
	case "real":
		return "REAL"
	case "money":
		return "DECIMAL(19,4)" // Standard money precision
	
	// Boolean type
	case "boolean":
		return "BOOLEAN"
	
	// Date and time types
	case "date":
		return "DATE"
	case "time":
		return "TIME"
	case "datetime":
		return "DATETIME"
	case "timestamp":
		return "TIMESTAMP"
	case "year":
		return "YEAR"
	
	// Binary types
	case "binary":
		return "BINARY(1)"
	case "varbinary":
		return "VARBINARY(255)"
	case "blob", "bytea":
		return "BLOB"
	case "tinyblob":
		return "TINYBLOB"
	case "mediumblob":
		return "MEDIUMBLOB"
	case "longblob":
		return "LONGBLOB"
	
	// JSON type (MySQL 5.7+)
	case "json", "jsonb":
		return "JSON"
	
	// Geometric types (MySQL spatial extensions)
	case "point":
		return "POINT"
	case "line", "linestring":
		return "LINESTRING"
	case "polygon":
		return "POLYGON"
	case "multipoint":
		return "MULTIPOINT"
	case "multilinestring":
		return "MULTILINESTRING"
	case "multipolygon":
		return "MULTIPOLYGON"
	case "geometry":
		return "GEOMETRY"
	case "geometrycollection":
		return "GEOMETRYCOLLECTION"
	
	// Bit type
	case "bit":
		return "BIT(1)"
	
	// Enum and Set types (simplified)
	case "enum":
		return "ENUM('value1', 'value2', 'value3')"
	case "set":
		return "SET('option1', 'option2', 'option3')"
	
	// Handle parameterized types
	default:
		lowerType := strings.ToLower(schemaType)
		if strings.HasPrefix(lowerType, "varchar") {
			return strings.ToUpper(schemaType)
		}
		if strings.HasPrefix(lowerType, "char") {
			return strings.ToUpper(schemaType)
		}
		if strings.HasPrefix(lowerType, "decimal") || strings.HasPrefix(lowerType, "numeric") {
			return strings.ToUpper(schemaType)
		}
		if strings.HasPrefix(lowerType, "binary") || strings.HasPrefix(lowerType, "varbinary") {
			return strings.ToUpper(schemaType)
		}
		if strings.HasPrefix(lowerType, "bit") {
			return strings.ToUpper(schemaType)
		}
		if strings.HasPrefix(lowerType, "enum") {
			return strings.ToUpper(schemaType)
		}
		if strings.HasPrefix(lowerType, "set") {
			return strings.ToUpper(schemaType)
		}
		
		log.Printf("Warning: Unknown schema type '%s', mapping to TEXT", schemaType)
		return "TEXT"
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

// GetAllIDs retrieves all id values from a table
func (m *MySQLDatabase) GetAllIDs(ctx context.Context, tableName string) ([]string, error) {
	query := fmt.Sprintf("SELECT id FROM %s", tableName)
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id sql.NullString
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan id: %w", err)
		}
		if id.Valid {
			ids = append(ids, id.String)
		}
	}
	return ids, nil
}

// GetAllForeignKeys retrieves all values for a specific foreign key column
func (m *MySQLDatabase) GetAllForeignKeys(ctx context.Context, tableName string, columnName string) ([]string, error) {
	query := fmt.Sprintf("SELECT %s FROM %s", columnName, tableName)
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table: %w", err)
	}
	defer rows.Close()

	var fks []string
	for rows.Next() {
		var fk sql.NullString
		if err := rows.Scan(&fk); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}
		if fk.Valid {
			fks = append(fks, fk.String)
		}
	}
	return fks, nil
}

// UpdateData updates existing data in a table
func (m *MySQLDatabase) UpdateData(ctx context.Context, tableName string, data []map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update each row
	for _, row := range data {
		// Extract the id field
		id, ok := row["id"]
		if !ok {
			return fmt.Errorf("document missing id field")
		}

		// Build SET clause
		var setClauses []string
		var args []interface{}
		argIndex := 1

		for col, val := range row {
			if col != "id" { // Skip id in SET clause
				setClauses = append(setClauses, fmt.Sprintf("`%s` = ?", col))
				args = append(args, val)
				argIndex++
			}
		}

		// Build and execute UPDATE query
		query := fmt.Sprintf("UPDATE `%s` SET %s WHERE `id` = ?",
			tableName,
			strings.Join(setClauses, ", "),
		)
		args = append(args, id)

		_, err := tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to update row: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// VerifyReferentialIntegrity checks if foreign key references are valid
func (m *MySQLDatabase) VerifyReferentialIntegrity(ctx context.Context, fromTable, fromColumn, toTable, toColumn string) error {
	log.Printf("Verifying referential integrity: %s.%s -> %s.%s", fromTable, fromColumn, toTable, toColumn)

	// Check for orphaned records in the 'toTable' (child) that reference non-existent records in 'fromTable' (parent)
	query := fmt.Sprintf(`
SELECT COUNT(*)
FROM %s t
LEFT JOIN %s f ON t.%s = f.%s
WHERE t.%s IS NOT NULL AND f.%s IS NULL`, // Count rows in child where FK is not null but parent does not exist
		"`"+toTable+"`", "`"+fromTable+"`", "`"+toColumn+"`", "`"+fromColumn+"`", "`"+toColumn+"`", "`"+fromColumn+"`",
	)

	log.Printf("Executing integrity check SQL: %s", query)

	var invalidCount int
	err := m.db.QueryRowContext(ctx, query).Scan(&invalidCount)
	if err != nil {
		return fmt.Errorf("failed to execute integrity check query: %w", err)
	}

	if invalidCount > 0 {
		return fmt.Errorf("found %d invalid foreign key references in %s.%s", invalidCount, toTable, toColumn)
	}

	log.Printf("Referential integrity check passed for %s.%s -> %s.%s", fromTable, fromColumn, toTable, toColumn)

	return nil
}

// Backup creates a backup of the database
func (m *MySQLDatabase) Backup(ctx context.Context, backupPath string) error {
	log.Printf("Creating backup of database '%s' to '%s'", m.config.Database, backupPath)

	// Execute mysqldump
	if err := exec.CommandContext(ctx, "mysqldump", "-h", m.config.Host, "-P", fmt.Sprintf("%d", m.config.Port),
		"-u", m.config.Username, "-p"+m.config.Password, "--single-transaction", "--quick", "--lock-tables=false",
		"--routines", "--triggers", "--events", m.config.Database, "-r", backupPath).Run(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	log.Printf("Backup created successfully at '%s'", backupPath)
	return nil
}

// Restore implements types.Database.
func (m *MySQLDatabase) Restore(ctx context.Context, backupFile string) error {
	// Use the mysql command-line client to restore the database
	// We need the database name and potentially connection details from the config
	dbName := m.config.Database

	// Construct the command. We'll read the backup file and pipe it to the mysql command.
	// This requires the mysql client to be in the PATH.

	// The command format is generally: mysql -h host -P port -u user -p database < backupfile
	// We need to extract host, port, user, and password from the DSN.
	// Parsing the DSN manually here is complex and error-prone.
	// A simpler approach for now is to rely on environment variables (like MYSQL_PWD) or configuration files.
	// Let's construct the basic command assuming connection details are handled externally or via a simple DSN.

	// Basic command: mysql --database dbname < backupfile

	// We need to run this by piping the file content to the mysql command's standard input
	// This is best done with exec.Command and StdIn.

	cmd := exec.CommandContext(ctx, "mysql", "--database", dbName)

	// TODO: Add logic to pass host, port, user, and password from DSN safely

	// Open the backup file
	backupFileReader, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file %s: %w", backupFile, err)
	}
	defer backupFileReader.Close()

	// Pipe the file content to the mysql command's standard input
	cmd.Stdin = backupFileReader

	// Run the command and capture output/errors
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute mysql restore: %v\nOutput: %s", err, string(output))
	}

	// Print output if any (might contain warnings or info)
	if len(output) > 0 {
		fmt.Printf("mysql restore output:\n%s\n", string(output))
	}

	return nil
}
