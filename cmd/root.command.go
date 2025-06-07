package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mockcraft",
	Short: "MockCraft - Universal Fake Data Generator",
	Long: `MockCraft is a comprehensive fake data toolkit with three modes:
- CLI faker for generating individual pieces of fake data
- Database seeder for populating databases with fake data
- REST API server for programmatic access to fake data generation`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(seedCmd)
	rootCmd.AddCommand(RestoreCmd)
}
