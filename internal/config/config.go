package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Output   OutputConfig   `yaml:"output"`
}

// ServerConfig represents server-specific configuration
type ServerConfig struct {
	Port        int      `yaml:"port" env:"MOCKCRAFT_PORT"`
	Host        string   `yaml:"host" env:"MOCKCRAFT_HOST"`
	RateLimit   int      `yaml:"rate_limit" env:"MOCKCRAFT_RATE_LIMIT"`
	CORSOrigins []string `yaml:"cors_origins" env:"MOCKCRAFT_CORS_ORIGINS"`
	MaxFileSize int64    `yaml:"max_file_size" env:"MOCKCRAFT_MAX_FILE_SIZE"`
	JobTimeout  int      `yaml:"job_timeout" env:"MOCKCRAFT_JOB_TIMEOUT"`
	JobCleanup  int      `yaml:"job_cleanup" env:"MOCKCRAFT_JOB_CLEANUP"`
}

// DatabaseConfig represents database connection settings
type DatabaseConfig struct {
	Type     string `yaml:"type" env:"MOCKCRAFT_DB_TYPE"`
	Host     string `yaml:"host" env:"MOCKCRAFT_DB_HOST"`
	Port     int    `yaml:"port" env:"MOCKCRAFT_DB_PORT"`
	User     string `yaml:"user" env:"MOCKCRAFT_DB_USER"`
	Password string `yaml:"password" env:"MOCKCRAFT_DB_PASSWORD"`
	Database string `yaml:"database" env:"MOCKCRAFT_DB_NAME"`
	SSLMode  string `yaml:"ssl_mode" env:"MOCKCRAFT_DB_SSL_MODE"`
}

// OutputConfig represents output generation settings
type OutputConfig struct {
	Format string `yaml:"format" env:"MOCKCRAFT_OUTPUT_FORMAT"`
	Dir    string `yaml:"dir" env:"MOCKCRAFT_OUTPUT_DIR"`
	Pretty bool   `yaml:"pretty" env:"MOCKCRAFT_OUTPUT_PRETTY"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := &Config{}

	// Load from file if path is provided
	if configPath != "" {
		if err := loadFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("error loading config file: %v", err)
		}
	}

	// Override with environment variables
	if err := loadFromEnv(config); err != nil {
		return nil, fmt.Errorf("error loading environment variables: %v", err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return config, nil
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(path string, config *Config) error {
	// Ensure file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", path)
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("error parsing config file: %v", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) error {
	// TODO: Implement environment variable loading
	return nil
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate server config
	if config.Server.Port <= 0 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	// Validate database config
	if config.Database.Type == "" {
		return fmt.Errorf("database type is required")
	}

	// Validate output config
	if config.Output.Format != "" {
		switch config.Output.Format {
		case "csv", "json", "sql":
			// Valid formats
		default:
			return fmt.Errorf("invalid output format: %s", config.Output.Format)
		}
	}

	return nil
}
