package types

import (
	"context"
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
}

// Column represents a database column
type Column struct {
	Name       string
	Type       string
	IsPrimary  bool
	IsNullable bool
	IsUnique   bool
	Default    interface{}
	Generator  string
	Params     map[string]interface{}
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

// Schema represents a database schema
type Schema struct {
	Tables []Table
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
	CreateTable(ctx context.Context, tableName string, columns []Column) error
	CreateIndex(ctx context.Context, tableName string, index Index) error
	InsertData(ctx context.Context, tableName string, data []map[string]interface{}) error
	BeginTransaction(ctx context.Context) (Transaction, error)
	GetDriver() string
}
