package database

import (
	"fmt"

	"github.com/valentinesamuel/mockcraft/internal/database/mongodb"
	// "github.com/valentinesamuel/mockcraft/internal/database/mysql"
	// "github.com/valentinesamuel/mockcraft/internal/database/postgres"
	// "github.com/valentinesamuel/mockcraft/internal/database/sqlite"
	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

// NewDatabase creates a new database instance based on the driver
func NewDatabase(config types.Config) (types.Database, error) {
	switch config.Driver {
	// case "postgres":
	// 	return postgres.NewPostgresDatabase(config)
	// case "mysql":
	// 	return mysql.NewMySQLDatabase(config)
	// case "sqlite":
	// 	return sqlite.NewSQLiteDatabase(config)
	case "mongodb":
		return mongodb.NewMongoDatabase(&config)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
}
