package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/valentinesamuel/mockcraft/internal/generators/base"
	"github.com/valentinesamuel/mockcraft/internal/ui"
)

func init() {
	rootCmd.AddCommand(generateCmd)

	// List flags
	generateCmd.Flags().BoolP("list", "l", false, "List all available types")
	generateCmd.Flags().BoolP("categories", "c", false, "List all categories")
	generateCmd.Flags().String("category", "", "Filter types by category")

	// Common parameters
	generateCmd.Flags().Int("length", 0, "Length for types that support it (e.g., password)")
	generateCmd.Flags().Int("word_count", 0, "Number of words for types that support it (e.g., sentence)")
	generateCmd.Flags().String("strings", "", "Comma-separated list of strings for shuffle_strings type")

	// Type-specific parameters
	generateCmd.Flags().String("phone_format", "", "Format for phone numbers (international, national, local)")
	generateCmd.Flags().Int("uuid_version", 0, "UUID version (1, 3, 4, 5)")
	generateCmd.Flags().String("tld", "", "Top-level domain (e.g., com, org, net)")
	generateCmd.Flags().Int("min_length", 0, "Minimum length for word generation")
	generateCmd.Flags().Int("max_length", 0, "Maximum length for word generation")
	generateCmd.Flags().Int("sentence_count", 0, "Number of sentences in paragraph")
	generateCmd.Flags().Float64("min", 0, "Minimum value for random number generation")
	generateCmd.Flags().Float64("max", 0, "Maximum value for random number generation")
	generateCmd.Flags().Int("precision", 0, "Precision for random float generation")
}

var generateCmd = &cobra.Command{
	Use:   "generate [type]",
	Short: "Generate fake data",
	Long: `Generate fake data of various types. For example:
  mockcraft generate first_name
  mockcraft generate password --length=16
  mockcraft generate sentence --word_count=10
  mockcraft generate phone --phone_format=international
  mockcraft generate address
  mockcraft generate uuid --uuid_version=4
  mockcraft generate domain --tld=org
  mockcraft generate word --min_length=5 --max_length=10
  mockcraft generate paragraph --sentence_count=5
  mockcraft generate random_int --min=1 --max=100
  mockcraft generate random_float --min=0.0 --max=1.0 --precision=3`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// List categories if requested
		if listCategories, _ := cmd.Flags().GetBool("categories"); listCategories {
			ui.PrintCategories()
			return nil
		}

		// List types in a category if requested
		if category, _ := cmd.Flags().GetString("category"); category != "" {
			ui.PrintTypesByCategory(category)
			return nil
		}

		// List all types if requested
		if listTypes, _ := cmd.Flags().GetBool("list"); listTypes {
			ui.PrintAllTypes()
			return nil
		}

		// Check if type is provided
		if len(args) == 0 {
			return fmt.Errorf("type is required")
		}

		// Get parameters from flags
		params := make(map[string]interface{})

		// Common parameters
		if length, _ := cmd.Flags().GetInt("length"); length > 0 {
			params["length"] = length
		}
		if wordCount, _ := cmd.Flags().GetInt("word_count"); wordCount > 0 {
			params["word_count"] = wordCount
		}
		if strings, _ := cmd.Flags().GetString("strings"); strings != "" {
			params["strings"] = strings
		}

		// Phone parameters
		if format, _ := cmd.Flags().GetString("phone_format"); format != "" {
			params["format"] = format
		}

		// UUID parameters
		if version, _ := cmd.Flags().GetInt("uuid_version"); version > 0 {
			params["version"] = version
		}

		// Domain parameters
		if tld, _ := cmd.Flags().GetString("tld"); tld != "" {
			params["tld"] = tld
		}

		// Word parameters
		if minLength, _ := cmd.Flags().GetInt("min_length"); minLength > 0 {
			params["min_length"] = minLength
		}
		if maxLength, _ := cmd.Flags().GetInt("max_length"); maxLength > 0 {
			params["max_length"] = maxLength
		}

		// Paragraph parameters
		if sentenceCount, _ := cmd.Flags().GetInt("sentence_count"); sentenceCount > 0 {
			params["sentence_count"] = sentenceCount
		}

		// Random number parameters
		if min, _ := cmd.Flags().GetFloat64("min"); min != 0 {
			params["min"] = min
		}
		if max, _ := cmd.Flags().GetFloat64("max"); max != 0 {
			params["max"] = max
		}
		if precision, _ := cmd.Flags().GetInt("precision"); precision > 0 {
			params["precision"] = precision
		}

		// Generate data
		generator := base.NewBaseGenerator()
		result, err := generator.GenerateByType(args[0], params)
		if err != nil {
			return fmt.Errorf("error generating data: %v", err)
		}

		// Print result in text format
		fmt.Println(result)
		return nil
	},
}
