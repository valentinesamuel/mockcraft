package database

import (
	"context"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

// Database represents a database connection
type Database interface {
	Connect(ctx context.Context) error
	Close() error
	CreateTable(ctx context.Context, tableName string, columns []types.Column) error
	CreateIndex(ctx context.Context, tableName string, index types.Index) error
	InsertData(ctx context.Context, tableName string, data []map[string]interface{}) error
	BeginTransaction(ctx context.Context) (types.Transaction, error)
	GetDriver() string
}
