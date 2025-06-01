package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/valentinesamuel/mockcraft/internal/generators"
)

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().BoolP("list", "l", false, "List all available types")
	generateCmd.Flags().Int("length", 12, "Length for password generation")
	generateCmd.Flags().Int("word_count", 6, "Word count for sentence generation")
	generateCmd.Flags().String("strings", "", "Comma-separated list of strings to shuffle")
}

var generateCmd = &cobra.Command{
	Use:   "generate [type]",
	Short: "Generate fake data of a specific type",
	Long: `Generate fake data of a specific type. For example:
mockcraft generate firstname    # Generate a first name
mockcraft generate blood_type  # Generate a blood type
mockcraft generate --list      # List all available types`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		list, _ := cmd.Flags().GetBool("list")
		if list {
			types := generators.GetAllAvailableTypes()
			for _, t := range types {
				fmt.Println(t)
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("please specify a type to generate or use --list to see available types")
		}

		dataType := args[0]
		params := make(map[string]interface{})

		// Get all flags and add them to params
		cmd.Flags().Visit(func(flag *pflag.Flag) {
			if flag.Name != "list" {
				params[flag.Name] = flag.Value.String()
			}
		})

		// Retrieve length and word_count flags and add them to params
		if length, err := cmd.Flags().GetInt("length"); err == nil {
			params["length"] = length
		}
		if wordCount, err := cmd.Flags().GetInt("word_count"); err == nil {
			params["word_count"] = wordCount
		}

		// Try to generate data using any available generator
		result, err := generators.Generate("", dataType, params)
		if err != nil {
			return err
		}

		// Pretty print the result as JSON
		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(jsonBytes))
		return nil
	},
}
