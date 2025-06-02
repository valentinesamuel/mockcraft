package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Test loading from file
	config, err := LoadConfig("configs/server.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify server config
	if config.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Server.Port)
	}
	if config.Server.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", config.Server.Host)
	}
	if config.Server.RateLimit != 100 {
		t.Errorf("Expected rate limit 100, got %d", config.Server.RateLimit)
	}

	// Verify database config
	if config.Database.Type != "postgres" {
		t.Errorf("Expected database type postgres, got %s", config.Database.Type)
	}
	if config.Database.Port != 5432 {
		t.Errorf("Expected database port 5432, got %d", config.Database.Port)
	}

	// Verify output config
	if config.Output.Format != "json" {
		t.Errorf("Expected output format json, got %s", config.Output.Format)
	}
	if !config.Output.Pretty {
		t.Error("Expected pretty output to be true")
	}
}

func TestLoadConfigWithEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("MOCKCRAFT_PORT", "9090")
	os.Setenv("MOCKCRAFT_DB_TYPE", "mysql")
	defer os.Unsetenv("MOCKCRAFT_PORT")
	defer os.Unsetenv("MOCKCRAFT_DB_TYPE")

	// Load config
	config, err := LoadConfig("configs/server.yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify environment variables override file config
	if config.Server.Port != 9090 {
		t.Errorf("Expected port 9090 from env, got %d", config.Server.Port)
	}
	if config.Database.Type != "mysql" {
		t.Errorf("Expected database type mysql from env, got %s", config.Database.Type)
	}
}

func TestValidateConfig(t *testing.T) {
	// Test invalid port
	config := &Config{
		Server: ServerConfig{
			Port: -1,
		},
	}
	if err := validateConfig(config); err == nil {
		t.Error("Expected error for invalid port")
	}

	// Test invalid database type
	config = &Config{
		Database: DatabaseConfig{
			Type: "",
		},
	}
	if err := validateConfig(config); err == nil {
		t.Error("Expected error for missing database type")
	}

	// Test invalid output format
	config = &Config{
		Output: OutputConfig{
			Format: "invalid",
		},
	}
	if err := validateConfig(config); err == nil {
		t.Error("Expected error for invalid output format")
	}
}
