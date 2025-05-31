package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed a database with fake data",
	Long: `Seed a database with fake data based on a YAML schema configuration.
Example:
mockcraft seed --config schema.yaml --db postgres://...`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement seeding logic
		fmt.Println("Seeding database with fake data...")
	},
}

func init() {
	rootCmd.AddCommand(seedCmd)
	seedCmd.Flags().String("config", "", "Path to schema configuration file")
	seedCmd.Flags().String("db", "", "Database connection string")
	seedCmd.Flags().String("output", "", "Output format (csv, json, sql)")
	seedCmd.Flags().String("dir", "", "Output directory for file-based output")
	seedCmd.Flags().Bool("backup", false, "Create database backup before seeding")
	seedCmd.Flags().String("backup-path", "", "Path to store database backups")
}
