package types

import (
	"context"
	"fmt"
	"time"
)

// Config represents database connection settings
type Config struct {
	Driver          string
	Host            string
	Port            int
	Username        string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	Options         map[string]string // Additional driver-specific options
}

// Column represents a database column
type Column struct {
	Name         string                 `yaml:"name"`
	Type         string                 `yaml:"type"`
	IsPrimary    bool                   `yaml:"is_primary,omitempty"`
	PrimaryKey   bool                   `yaml:"primary_key,omitempty"` // Alternative naming for MongoDB
	IsNullable   bool                   `yaml:"is_nullable,omitempty" default:"false"`
	IsUnique     bool                   `yaml:"is_unique,omitempty"`
	Default      interface{}            `yaml:"default,omitempty"`
	Generator    string                 `yaml:"generator,omitempty"`
	Industry     string                 `yaml:"industry,omitempty"`
	Params       map[string]interface{} `yaml:"params,omitempty"`
	Values       []string               `yaml:"values,omitempty"` // For enum generator
	IsForeign    bool                   `yaml:"is_foreign,omitempty"`
	NestedFields []Column               `yaml:"nested_fields,omitempty"` // For MongoDB embedded documents
	Subtype      string                 `yaml:"subtype,omitempty"`       // For MongoDB binary subtypes
}

// Index represents a database index
type Index struct {
	Name       string
	Columns    []string
	IsUnique   bool
	Type       string
	Properties map[string]interface{}
}

// Table represents a database table
type Table struct {
	Name    string
	Columns []Column
	Indexes []Index
	Data    []map[string]interface{}
	Count   int
}

// GetPrimaryKeyColumn returns the primary key column of the table
func (t *Table) GetPrimaryKeyColumn() (*Column, error) {
	for _, col := range t.Columns {
		if col.IsPrimary {
			return &col, nil
		}
	}
	return nil, fmt.Errorf("primary key column not found for table %s", t.Name)
}

// Relationship represents a relationship between tables
type Relationship struct {
	Type       string `yaml:"type"`
	FromTable  string `yaml:"from_table"`
	FromColumn string `yaml:"from_column"`
	ToTable    string `yaml:"to_table"`
	ToColumn   string `yaml:"to_column"`
}

// Constraint represents a database constraint
type Constraint struct {
	Type      string // foreign_key, unique, etc.
	Columns   []string
	Condition string
}

// Schema represents a database schema
type Schema struct {
	Tables      []Table
	Relations   []Relationship
	Constraints []Constraint
	// MongoDB alternative format
	Collections []Collection `yaml:"collections,omitempty"`
}

// Collection represents a MongoDB collection (alternative to Table)
type Collection struct {
	Name   string   `yaml:"name"`
	Count  int      `yaml:"count"`
	Fields []Column `yaml:"fields"`
}

// Transaction represents a database transaction
type Transaction interface {
	Commit() error
	Rollback() error
}

// Database represents a database connection
type Database interface {
	Connect(ctx context.Context) error
	Close() error
	CreateTable(ctx context.Context, tableName string, table *Table, relations []Relationship) error
	CreateIndex(ctx context.Context, tableName string, index Index) error
	CreateConstraint(ctx context.Context, tableName string, constraint Constraint) error
	InsertData(ctx context.Context, tableName string, data []map[string]interface{}) error
	UpdateData(ctx context.Context, tableName string, data []map[string]interface{}) error
	GetAllIDs(ctx context.Context, tableName string) ([]string, error)
	GetAllForeignKeys(ctx context.Context, tableName string, columnName string) ([]string, error)
	VerifyReferentialIntegrity(ctx context.Context, fromTable, fromColumn, toTable, toColumn string) error
	DropTable(ctx context.Context, tableName string) error
	GetDriver() string
	BeginTransaction(ctx context.Context) (Transaction, error)
	Backup(ctx context.Context, backupPath string) error
	Restore(ctx context.Context, backupPath string) error
}
