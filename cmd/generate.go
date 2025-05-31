package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(generateCmd)
  }

var generateCmd = &cobra.Command{
	Use:   "generate [type]",
	Short: "Generate fake data of a specific type",
	Long: `Generate fake data of a specific type. For example:
mockcraft generate firstname    # Generate a fake first name
mockcraft generate password     # Generate a fake password`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement generation logic
	},
}

