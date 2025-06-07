package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"

	// Import the database package from internal/database
	"github.com/valentinesamuel/mockcraft/internal/database"
)

var (
	dbDSN          string
	backupFilePath string
)

// RestoreCmd represents the restore command
var RestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore database from a backup file",
	Long:  `Restore database from a backup file for various database types.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flag values
		dbDSN, _ := cmd.Flags().GetString("db")
		backupFilePath, _ := cmd.Flags().GetString("backup-file")

		if dbDSN == "" {
			return fmt.Errorf("--db flag is required")
		}
		if backupFilePath == "" {
			return fmt.Errorf("--backup-file flag is required")
		}

		// Parse the database URL using the database package
		config, err := database.ParseDatabaseURL(dbDSN)
		if err != nil {
			return fmt.Errorf("failed to parse database DSN: %w", err)
		}

		// Create database instance
		db, err := database.NewDatabase(config)
		if err != nil {
			return fmt.Errorf("failed to create database instance: %w", err)
		}
		defer db.Close()

		// Establish database connection
		ctx := context.Background()
		if err := db.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		// Perform restoration
		log.Printf("Starting database restoration from %s", backupFilePath)
		if err := db.Restore(ctx, backupFilePath); err != nil {
			return fmt.Errorf("database restoration failed: %w", err)
		}

		log.Println("Database restoration completed successfully.")

		return nil
	},
}

func init() {
	// Add flags to the restore command
	RestoreCmd.Flags().StringVarP(&dbDSN, "db", "d", "", "Database connection string (DSN)")
	RestoreCmd.Flags().StringVarP(&backupFilePath, "backup-file", "f", "", "Path to the backup file")

	// We will add this command to the root command in root.command.go
}
