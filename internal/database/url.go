package database

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
)

// ParseDatabaseURL parses a database URL into a database configuration
func ParseDatabaseURL(dsn string) (types.Config, error) {
	// Handle SQLite special case
	if strings.HasPrefix(dsn, "sqlite://") {
		return parseSQLiteURL(dsn)
	}

	// Handle MongoDB special case
	if strings.HasPrefix(dsn, "mongodb://") || strings.HasPrefix(dsn, "mongodb+srv://") {
		return parseMongoURL(dsn)
	}

	// Parse URL
	u, err := url.Parse(dsn)
	if err != nil {
		return types.Config{}, fmt.Errorf("invalid database URL: %w", err)
	}

	// Get driver from scheme
	driver := strings.TrimSuffix(u.Scheme, "ql") // postgresql -> postgres, mysql -> mysql
	if driver != "postgres" && driver != "mysql" {
		return types.Config{}, fmt.Errorf("unsupported database driver: %s", driver)
	}

	// Parse host and port
	host := u.Hostname()
	port := 0
	if u.Port() != "" {
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return types.Config{}, fmt.Errorf("invalid port number: %w", err)
		}
	} else {
		// Set default ports
		switch driver {
		case "postgres":
			port = 5432
		case "mysql":
			port = 3306
		}
	}

	// Get database name from path
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return types.Config{}, fmt.Errorf("database name is required")
	}

	// Get username and password
	username := u.User.Username()
	password, _ := u.User.Password()

	// Parse query parameters
	query := u.Query()
	sslMode := query.Get("sslmode")
	if sslMode == "" {
		sslMode = "disable" // Default to disable SSL
	}

	// Parse connection pool settings
	maxOpenConns := 10
	if v := query.Get("max_open_conns"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxOpenConns = n
		}
	}

	maxIdleConns := 5
	if v := query.Get("max_idle_conns"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxIdleConns = n
		}
	}

	connMaxLifetime := time.Hour
	if v := query.Get("conn_max_lifetime"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			connMaxLifetime = d
		}
	}

	connMaxIdleTime := 5 * time.Minute
	if v := query.Get("conn_max_idle_time"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			connMaxIdleTime = d
		}
	}

	return types.Config{
		Driver:          driver,
		Host:            host,
		Port:            port,
		Username:        username,
		Password:        password,
		Database:        dbName,
		SSLMode:         sslMode,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
	}, nil
}

// parseMongoURL parses a MongoDB connection URL
func parseMongoURL(dsn string) (types.Config, error) {
	// Parse URL
	u, err := url.Parse(dsn)
	if err != nil {
		return types.Config{}, fmt.Errorf("invalid MongoDB URL: %w", err)
	}

	// Get username and password
	username := u.User.Username()
	password, _ := u.User.Password()

	// Get database name from path
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return types.Config{}, fmt.Errorf("database name is required")
	}

	// Parse host and port
	host := u.Hostname()
	port := 27017 // Default MongoDB port
	if u.Port() != "" {
		var err error
		port, err = strconv.Atoi(u.Port())
		if err != nil {
			return types.Config{}, fmt.Errorf("invalid port number: %w", err)
		}
	}

	// Parse query parameters
	query := u.Query()

	// Parse connection pool settings
	maxOpenConns := 100 // MongoDB default is higher
	if v := query.Get("maxPoolSize"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxOpenConns = n
		}
	}

	maxIdleConns := 100
	if v := query.Get("minPoolSize"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxIdleConns = n
		}
	}

	connMaxLifetime := time.Hour
	if v := query.Get("maxIdleTimeMS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			connMaxLifetime = time.Duration(n) * time.Millisecond
		}
	}

	connMaxIdleTime := 5 * time.Minute
	if v := query.Get("maxIdleTimeMS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			connMaxIdleTime = time.Duration(n) * time.Millisecond
		}
	}

	// MongoDB specific options
	sslMode := "disable"
	if query.Get("ssl") == "true" || query.Get("tls") == "true" {
		sslMode = "require"
	}

	// Handle replica set and other MongoDB specific options
	replicaSet := query.Get("replicaSet")
	authSource := query.Get("authSource")
	if authSource == "" {
		authSource = "admin" // Default auth source
	}

	// Build connection string
	connStr := dsn
	if replicaSet != "" {
		if !strings.Contains(connStr, "replicaSet=") {
			connStr += "&replicaSet=" + replicaSet
		}
	}
	if authSource != "" && !strings.Contains(connStr, "authSource=") {
		connStr += "&authSource=" + authSource
	}

	return types.Config{
		Driver:          "mongodb",
		Host:            host,
		Port:            port,
		Username:        username,
		Password:        password,
		Database:        dbName,
		SSLMode:         sslMode,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
		Options: map[string]string{
			"replicaSet": replicaSet,
			"authSource": authSource,
		},
	}, nil
}

// parseSQLiteURL parses a SQLite database URL
func parseSQLiteURL(dsn string) (types.Config, error) {
	// Remove sqlite:// prefix
	path := strings.TrimPrefix(dsn, "sqlite://")
	if path == "" {
		return types.Config{}, fmt.Errorf("SQLite database path is required")
	}

	// Parse query parameters
	u, err := url.Parse(dsn)
	if err != nil {
		return types.Config{}, fmt.Errorf("invalid SQLite URL: %w", err)
	}

	query := u.Query()
	maxOpenConns := 1 // SQLite only supports one connection at a time
	if v := query.Get("max_open_conns"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxOpenConns = n
		}
	}

	maxIdleConns := 1
	if v := query.Get("max_idle_conns"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxIdleConns = n
		}
	}

	connMaxLifetime := time.Hour
	if v := query.Get("conn_max_lifetime"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			connMaxLifetime = d
		}
	}

	connMaxIdleTime := 5 * time.Minute
	if v := query.Get("conn_max_idle_time"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			connMaxIdleTime = d
		}
	}

	return types.Config{
		Driver:          "sqlite3",
		Database:        path,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
		ConnMaxIdleTime: connMaxIdleTime,
	}, nil
}
