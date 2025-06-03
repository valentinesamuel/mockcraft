package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/valentinesamuel/mockcraft/internal/database/types"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig `yaml:"server"`
	Database types.Config `yaml:"database"`
	Output   OutputConfig `yaml:"output"`
}

// ServerConfig represents server settings
type ServerConfig struct {
	Host         string `yaml:"host" env:"MOCKCRAFT_HOST"`
	Port         int    `yaml:"port" env:"MOCKCRAFT_PORT"`
	ReadTimeout  int    `yaml:"read_timeout" env:"MOCKCRAFT_READ_TIMEOUT"`
	WriteTimeout int    `yaml:"write_timeout" env:"MOCKCRAFT_WRITE_TIMEOUT"`
	JobCleanup   int    `yaml:"job_cleanup" env:"MOCKCRAFT_JOB_CLEANUP"`
}

// OutputConfig represents output generation settings
type OutputConfig struct {
	Format string `yaml:"format" env:"MOCKCRAFT_OUTPUT_FORMAT"`
	Pretty bool   `yaml:"pretty" env:"MOCKCRAFT_OUTPUT_PRETTY"`
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	config := &Config{
		Database: types.Config{
			Driver:          getEnv("DB_DRIVER", "postgres"),
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			Username:        getEnv("DB_USERNAME", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Database:        getEnv("DB_NAME", "mockcraft"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// getEnvAsDuration gets an environment variable as a duration or returns a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return duration
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(path string) (*Config, error) {
	// Load from file
	config := &Config{}
	if err := loadFromFile(path, config); err != nil {
		return nil, err
	}

	// Load from environment variables
	if err := loadFromEnv(config); err != nil {
		return nil, err
	}

	// Validate config
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) error {
	// Load server config
	config.Server.Host = getEnv("MOCKCRAFT_HOST", config.Server.Host)
	config.Server.Port = getEnvAsInt("MOCKCRAFT_PORT", config.Server.Port)
	config.Server.ReadTimeout = getEnvAsInt("MOCKCRAFT_READ_TIMEOUT", config.Server.ReadTimeout)
	config.Server.WriteTimeout = getEnvAsInt("MOCKCRAFT_WRITE_TIMEOUT", config.Server.WriteTimeout)
	config.Server.JobCleanup = getEnvAsInt("MOCKCRAFT_JOB_CLEANUP", config.Server.JobCleanup)

	// Load database config
	config.Database.Driver = getEnv("DB_DRIVER", config.Database.Driver)
	config.Database.Host = getEnv("DB_HOST", config.Database.Host)
	config.Database.Port = getEnvAsInt("DB_PORT", config.Database.Port)
	config.Database.Username = getEnv("DB_USERNAME", config.Database.Username)
	config.Database.Password = getEnv("DB_PASSWORD", config.Database.Password)
	config.Database.Database = getEnv("DB_NAME", config.Database.Database)
	config.Database.SSLMode = getEnv("DB_SSL_MODE", config.Database.SSLMode)
	config.Database.MaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", config.Database.MaxOpenConns)
	config.Database.MaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", config.Database.MaxIdleConns)
	config.Database.ConnMaxLifetime = getEnvAsDuration("DB_CONN_MAX_LIFETIME", config.Database.ConnMaxLifetime)
	config.Database.ConnMaxIdleTime = getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", config.Database.ConnMaxIdleTime)

	// Load output config
	config.Output.Format = getEnv("MOCKCRAFT_OUTPUT_FORMAT", config.Output.Format)
	config.Output.Pretty = getEnvAsBool("MOCKCRAFT_OUTPUT_PRETTY", config.Output.Pretty)

	return nil
}

// getEnvAsBool gets an environment variable as a boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate server config
	if config.Server.Host == "" {
		return fmt.Errorf("server host is required")
	}
	if config.Server.Port <= 0 {
		return fmt.Errorf("server port must be greater than 0")
	}
	if config.Server.ReadTimeout <= 0 {
		return fmt.Errorf("server read timeout must be greater than 0")
	}
	if config.Server.WriteTimeout <= 0 {
		return fmt.Errorf("server write timeout must be greater than 0")
	}
	if config.Server.JobCleanup <= 0 {
		return fmt.Errorf("server job cleanup must be greater than 0")
	}

	// Validate database config
	if config.Database.Driver == "" {
		return fmt.Errorf("database driver is required")
	}
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.Port <= 0 {
		return fmt.Errorf("database port must be greater than 0")
	}
	if config.Database.Username == "" {
		return fmt.Errorf("database username is required")
	}
	if config.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}
	if config.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if config.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("database max open connections must be greater than 0")
	}
	if config.Database.MaxIdleConns <= 0 {
		return fmt.Errorf("database max idle connections must be greater than 0")
	}
	if config.Database.ConnMaxLifetime <= 0 {
		return fmt.Errorf("database connection max lifetime must be greater than 0")
	}
	if config.Database.ConnMaxIdleTime <= 0 {
		return fmt.Errorf("database connection max idle time must be greater than 0")
	}

	// Validate output config
	if config.Output.Format == "" {
		return fmt.Errorf("output format is required")
	}

	return nil
}
