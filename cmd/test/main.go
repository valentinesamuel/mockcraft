package main

import (
	"fmt"
	"log"

	"github.com/valentinesamuel/mockcraft/internal/config"
)

func main() {
	// Load server configuration
	cfg, err := config.LoadConfig("configs/server.yaml")
	if err != nil {
		log.Fatalf("Failed to load server config: %v", err)
	}

	// Print server configuration
	fmt.Println("=== Server Configuration ===")
	fmt.Printf("Port: %d\n", cfg.Server.Port)
	fmt.Printf("Host: %s\n", cfg.Server.Host)
	fmt.Printf("Rate Limit: %d\n", cfg.Server.RateLimit)
	fmt.Printf("CORS Origins: %v\n", cfg.Server.CORSOrigins)
	fmt.Printf("Max File Size: %d\n", cfg.Server.MaxFileSize)
	fmt.Printf("Job Timeout: %d\n", cfg.Server.JobTimeout)
	fmt.Printf("Job Cleanup: %d\n", cfg.Server.JobCleanup)

	// Print database configuration
	fmt.Println("\n=== Database Configuration ===")
	fmt.Printf("Type: %s\n", cfg.Database.Type)
	fmt.Printf("Host: %s\n", cfg.Database.Host)
	fmt.Printf("Port: %d\n", cfg.Database.Port)
	fmt.Printf("User: %s\n", cfg.Database.User)
	fmt.Printf("Database: %s\n", cfg.Database.Database)
	fmt.Printf("SSL Mode: %s\n", cfg.Database.SSLMode)

	// Print output configuration
	fmt.Println("\n=== Output Configuration ===")
	fmt.Printf("Format: %s\n", cfg.Output.Format)
	fmt.Printf("Directory: %s\n", cfg.Output.Dir)
	fmt.Printf("Pretty: %v\n", cfg.Output.Pretty)

	// Load and validate e-commerce schema
	schema, err := config.LoadSchema("configs/examples/ecommerce.yaml")
	if err != nil {
		log.Fatalf("Failed to load schema: %v", err)
	}

	// Print schema information
	fmt.Println("\n=== Schema Information ===")
	fmt.Printf("Number of tables: %d\n", len(schema.Tables))
	for _, table := range schema.Tables {
		fmt.Printf("\nTable: %s\n", table.Name)
		fmt.Printf("  Count: %d\n", table.Count)
		fmt.Printf("  Columns: %d\n", len(table.Columns))
		fmt.Printf("  Relations: %d\n", len(table.Relations))
		fmt.Printf("  Constraints: %d\n", len(table.Constraints))
	}
}
