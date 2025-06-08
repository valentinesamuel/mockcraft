package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/valentinesamuel/mockcraft/internal/generators/industries/base"
	"github.com/valentinesamuel/mockcraft/internal/generators/interfaces"
	"github.com/valentinesamuel/mockcraft/internal/generators/types"
	"github.com/valentinesamuel/mockcraft/internal/ui"
)

func init() {
	rootCmd.AddCommand(generateCmd)

	// List flags
	generateCmd.Flags().BoolP("list", "l", false, "List all available types")
	generateCmd.Flags().BoolP("categories", "c", false, "List all categories")
	generateCmd.Flags().String("category", "", "Filter types by category")

	// Batch generation flags
	generateCmd.Flags().IntP("count", "n", 1, "Number of values to generate")
	generateCmd.Flags().Int64("seed", 0, "Seed for random number generation (for reproducible results)")
	generateCmd.Flags().Bool("parallel", false, "Generate values in parallel")
	generateCmd.Flags().Int("workers", 4, "Number of parallel workers (when --parallel is used)")

	// Output format flags
	generateCmd.Flags().StringP("output", "o", "", "Output file path (default: stdout)")
	generateCmd.Flags().StringP("output-format", "f", "text", "Output format (text, json, pretty)")
	generateCmd.Flags().String("template", "", "Custom template for output formatting")
	generateCmd.Flags().Bool("pretty", false, "Pretty print the output (for JSON)")

	// Common parameters
	generateCmd.Flags().Int("length", 0, "Length for types that support it (e.g., password)")
	generateCmd.Flags().Int("word_count", 0, "Number of words for types that support it (e.g., sentence)")
	generateCmd.Flags().String("strings", "", "Comma-separated list of strings for shuffle_strings type")
	generateCmd.Flags().String("country", "", "Country code for location-specific data (e.g., US, GB, DE)")
	generateCmd.Flags().String("language", "", "Language code for text generation (e.g., en, es, fr)")
	generateCmd.Flags().String("format", "", "Format string for date/time values (e.g., 2006-01-02, 15:04:05)")

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

	// Date/Time parameters
	generateCmd.Flags().String("start_date", "", "Start date for date range (format: 2006-01-02)")
	generateCmd.Flags().String("end_date", "", "End date for date range (format: 2006-01-02)")
	generateCmd.Flags().String("timezone", "", "Timezone for date/time generation (e.g., UTC, America/New_York)")

	// Text parameters
	generateCmd.Flags().Bool("capitalize", false, "Capitalize the first letter of each word")
	generateCmd.Flags().Bool("lowercase", false, "Convert text to lowercase")
	generateCmd.Flags().Bool("uppercase", false, "Convert text to uppercase")
	generateCmd.Flags().String("prefix", "", "Add prefix to generated text")
	generateCmd.Flags().String("suffix", "", "Add suffix to generated text")

	// Number parameters
	generateCmd.Flags().Bool("positive", false, "Generate only positive numbers")
	generateCmd.Flags().Bool("negative", false, "Generate only negative numbers")
	generateCmd.Flags().Bool("integer", false, "Generate only integer values")
	generateCmd.Flags().String("unit", "", "Unit for numeric values (e.g., kg, m, $)")

	// Set custom help template
	generateCmd.SetHelpTemplate(`Usage:
  mockcraft generate [type] [flags]

Available Commands:
  list, l     List all available types
  categories, c  List all categories
  category    Filter types by category

Examples:
  mockcraft generate first_name
  mockcraft generate password --length=16
  mockcraft generate sentence --word_count=10
  mockcraft generate phone --phone_format=international --country=US
  mockcraft generate address --country=GB
  mockcraft generate uuid --uuid_version=4
  mockcraft generate domain --tld=org
  mockcraft generate word --min_length=5 --max_length=10
  mockcraft generate paragraph --sentence_count=5 --language=es
  mockcraft generate random_int --min=1 --max=100 --positive
  mockcraft generate random_float --min=0.0 --max=1.0 --precision=3 --unit=kg
  mockcraft generate date --start_date=2024-01-01 --end_date=2024-12-31 --format=2006-01-02
  mockcraft generate text --capitalize --prefix=Mr. --suffix=!

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

Use "mockcraft generate [type] --help" for more information about a specific type.`)

	// Override the default help command
	generateCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			typeDef := types.GetTypeByName(args[0])
			if typeDef != nil {
				fmt.Printf("Type: %s\n", typeDef.Name)
				fmt.Printf("Description: %s\n", typeDef.Description)
				fmt.Printf("Example: %s\n\n", typeDef.Example)

				if len(typeDef.Parameters) > 0 {
					fmt.Println("Available parameters:")
					for _, param := range typeDef.Parameters {
						required := ""
						if param.Required {
							required = " (required)"
						}
						fmt.Printf("  --%s%s: %s\n", param.Name, required, param.Description)
						if param.Default != nil {
							fmt.Printf("    Default: %v\n", param.Default)
						}
					}
				}
				return
			}
		}
		cmd.Usage()
	})
}

var generateCmd = &cobra.Command{
	Use:   "generate [type]",
	Short: "Generate fake data",
	Long: `Generate fake data of various types. For example:
  mockcraft generate first_name
  mockcraft generate password --length=16
  mockcraft generate sentence --word_count=10
  mockcraft generate phone --phone_format=international --country=US
  mockcraft generate address --country=GB
  mockcraft generate uuid --uuid_version=4
  mockcraft generate domain --tld=org
  mockcraft generate word --min_length=5 --max_length=10
  mockcraft generate paragraph --sentence_count=5 --language=es
  mockcraft generate random_int --min=1 --max=100 --positive
  mockcraft generate random_float --min=0.0 --max=1.0 --precision=3 --unit=kg
  mockcraft generate date --start_date=2024-01-01 --end_date=2024-12-31 --format=2006-01-02
  mockcraft generate text --capitalize --prefix=Mr. --suffix=!

Batch Generation:
  -n, --count int      Number of values to generate
  --seed int64         Seed for random number generation
  --parallel           Generate values in parallel
  --workers int        Number of parallel workers (default 4)

Output Options:
  -o, --output string        Output file path (default: stdout)
  -f, --output-format string Output format (text, json, pretty)
  --template string          Custom template for output formatting
  --pretty                   Pretty print the output (for JSON)`,
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

		// Get batch generation options
		count, _ := cmd.Flags().GetInt("count")
		seed, _ := cmd.Flags().GetInt64("seed")
		parallel, _ := cmd.Flags().GetBool("parallel")
		workers, _ := cmd.Flags().GetInt("workers")

		// Get output format options
		outputFile, _ := cmd.Flags().GetString("output")
		outputFormat, _ := cmd.Flags().GetString("output-format")
		customTemplate, _ := cmd.Flags().GetString("template")
		prettyPrint, _ := cmd.Flags().GetBool("pretty")

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
		if country, _ := cmd.Flags().GetString("country"); country != "" {
			params["country"] = strings.ToUpper(country)
		}
		if language, _ := cmd.Flags().GetString("language"); language != "" {
			params["language"] = strings.ToLower(language)
		}
		if format, _ := cmd.Flags().GetString("format"); format != "" {
			params["format"] = format
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

		// Date/Time parameters
		if startDate, _ := cmd.Flags().GetString("start_date"); startDate != "" {
			params["start_date"] = startDate
		}
		if endDate, _ := cmd.Flags().GetString("end_date"); endDate != "" {
			params["end_date"] = endDate
		}
		if timezone, _ := cmd.Flags().GetString("timezone"); timezone != "" {
			params["timezone"] = timezone
		}

		// Text parameters
		if capitalize, _ := cmd.Flags().GetBool("capitalize"); capitalize {
			params["capitalize"] = true
		}
		if lowercase, _ := cmd.Flags().GetBool("lowercase"); lowercase {
			params["lowercase"] = true
		}
		if uppercase, _ := cmd.Flags().GetBool("uppercase"); uppercase {
			params["uppercase"] = true
		}
		if prefix, _ := cmd.Flags().GetString("prefix"); prefix != "" {
			params["prefix"] = prefix
		}
		if suffix, _ := cmd.Flags().GetString("suffix"); suffix != "" {
			params["suffix"] = suffix
		}

		// Number parameters
		if positive, _ := cmd.Flags().GetBool("positive"); positive {
			params["positive"] = true
		}
		if negative, _ := cmd.Flags().GetBool("negative"); negative {
			params["negative"] = true
		}
		if integer, _ := cmd.Flags().GetBool("integer"); integer {
			params["integer"] = true
		}
		if unit, _ := cmd.Flags().GetString("unit"); unit != "" {
			params["unit"] = unit
		}

		// Validate parameter dependencies
		if err := validateParameterDependencies(params); err != nil {
			return err
		}

		// Generate data
		generator := base.NewBaseGenerator()
		if seed != 0 {
			generator.SetSeed(seed)
		}

		var results []interface{}
		if parallel {
			results = generateParallel(generator, args[0], params, count, workers)
		} else {
			results = generateSequential(generator, args[0], params, count)
		}

		// Format and output the results
		output, err := formatBatchOutput(results, outputFormat, customTemplate, prettyPrint)
		if err != nil {
			return fmt.Errorf("error formatting output: %v", err)
		}

		// Write to file or stdout
		if outputFile != "" {
			if err := writeToFile(outputFile, output); err != nil {
				return fmt.Errorf("error writing to file: %v", err)
			}
		} else {
			fmt.Println(output)
		}

		return nil
	},
}

// validateParameterDependencies checks for parameter dependencies and conflicts
func validateParameterDependencies(params map[string]interface{}) error {
	// Check for conflicting text case options
	if params["capitalize"] == true && (params["lowercase"] == true || params["uppercase"] == true) {
		return fmt.Errorf("cannot use --capitalize with --lowercase or --uppercase")
	}
	if params["lowercase"] == true && params["uppercase"] == true {
		return fmt.Errorf("cannot use both --lowercase and --uppercase")
	}

	// Check for conflicting number options
	if params["positive"] == true && params["negative"] == true {
		return fmt.Errorf("cannot use both --positive and --negative")
	}

	// Check for date range
	if params["start_date"] != nil && params["end_date"] != nil {
		startDate := params["start_date"].(string)
		endDate := params["end_date"].(string)
		if startDate > endDate {
			return fmt.Errorf("start_date must be before end_date")
		}
	}

	// Check for number range
	if params["min"] != nil && params["max"] != nil {
		min := params["min"].(float64)
		max := params["max"].(float64)
		if min > max {
			return fmt.Errorf("min value must be less than max value")
		}
		if params["positive"] == true && min < 0 {
			return fmt.Errorf("min value must be positive when using --positive")
		}
		if params["negative"] == true && max > 0 {
			return fmt.Errorf("max value must be negative when using --negative")
		}
	}

	// Check for length range
	if params["min_length"] != nil && params["max_length"] != nil {
		minLength := params["min_length"].(int)
		maxLength := params["max_length"].(int)
		if minLength > maxLength {
			return fmt.Errorf("min_length must be less than max_length")
		}
	}

	return nil
}

// generateSequential generates values sequentially
func generateSequential(generator interfaces.Generator, dataType string, params map[string]interface{}, count int) []interface{} {
	results := make([]interface{}, count)

	// Create progress bar
	progress := ui.NewProgressBar(count)
	progress.Start()

	// Print a newline to make room for results
	fmt.Println()

	for i := 0; i < count; i++ {
		result, err := generator.GenerateByType(dataType, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating value %d: %v\n", i+1, err)
			continue
		}
		results[i] = result
		// Print result immediately
		if str, ok := result.(string); ok {
			fmt.Println(str)
		} else if str, ok := result.(fmt.Stringer); ok {
			fmt.Println(str.String())
		} else {
			fmt.Printf("%v\n", result)
		}
		progress.Increment()
	}

	progress.Stop()
	return results
}

// generateParallel generates values in parallel
func generateParallel(generator interfaces.Generator, dataType string, params map[string]interface{}, count, workers int) []interface{} {
	results := make([]interface{}, count)
	var wg sync.WaitGroup
	jobs := make(chan int, count)
	resultsChan := make(chan struct {
		index int
		value interface{}
		err   error
	}, count)

	// Create progress bar
	progress := ui.NewProgressBar(count)
	progress.Start()

	// Print a newline to make room for results
	fmt.Println()

	// Create worker pool
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for jobIndex := range jobs {
				value, err := generator.GenerateByType(dataType, params)
				resultsChan <- struct {
					index int
					value interface{}
					err   error
				}{jobIndex, value, err}
			}
		}()
	}

	// Submit jobs
	for i := 0; i < count; i++ {
		jobs <- i
	}
	close(jobs)

	// Collect results
	for i := 0; i < count; i++ {
		result := <-resultsChan
		if result.err != nil {
			fmt.Fprintf(os.Stderr, "Error generating value %d: %v\n", result.index+1, result.err)
			continue
		}
		results[result.index] = result.value
		// Print result immediately
		if str, ok := result.value.(string); ok {
			fmt.Println(str)
		} else if str, ok := result.value.(fmt.Stringer); ok {
			fmt.Println(str.String())
		} else {
			 fmt.Printf("%v\n", result.value)
		}
		progress.Increment()
	}

	progress.Stop()
	wg.Wait()
	return results
}

// formatBatchOutput formats a batch of results
func formatBatchOutput(results []interface{}, format string, template string, pretty bool) (string, error) {
	switch format {
	case "json":
		return formatJSON(results, pretty)
	case "pretty":
		return formatPretty(results)
	case "text":
		if template != "" {
			return formatBatchWithTemplate(results, template)
		}
		return formatBatchText(results)
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

// formatBatchText formats a batch of results as plain text
func formatBatchText(results []interface{}) (string, error) {
	var lines []string
	for _, result := range results {
		switch v := result.(type) {
		case string:
			lines = append(lines, v)
		case fmt.Stringer:
			lines = append(lines, v.String())
		default:
			lines = append(lines, fmt.Sprintf("%v", v))
		}
	}
	return strings.Join(lines, "\n"), nil
}

// formatBatchWithTemplate formats a batch of results using a custom template
func formatBatchWithTemplate(results []interface{}, templateStr string) (string, error) {
	tmpl, err := template.New("output").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %v", err)
	}

	var lines []string
	for _, result := range results {
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, result); err != nil {
			return "", fmt.Errorf("error executing template: %v", err)
		}
		lines = append(lines, buf.String())
	}
	return strings.Join(lines, "\n"), nil
}

// formatJSON formats the result as JSON
func formatJSON(result interface{}, pretty bool) (string, error) {
	var jsonData []byte
	var err error

	if pretty {
		jsonData, err = json.MarshalIndent(result, "", "  ")
	} else {
		jsonData, err = json.Marshal(result)
	}

	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %v", err)
	}

	return string(jsonData), nil
}

// formatPretty formats the result in a pretty, human-readable format
func formatPretty(result interface{}) (string, error) {
	switch v := result.(type) {
	case string:
		return v, nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		// For complex types, use JSON pretty printing
		return formatJSON(result, true)
	}
}

// writeToFile writes the output to a file
func writeToFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
