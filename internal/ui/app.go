package ui

import (
	"fmt"
	"strings"

	"github.com/valentinesamuel/mockcraft/internal/generators"
)

// PrintCategories prints all available industries
func PrintCategories() {
	engine := generators.GetGlobalEngine()
	industries := engine.ListIndustries()
	fmt.Println("Available industries:")
	for _, industry := range industries {
		fmt.Printf("  %s\n", industry)
	}
}

// PrintTypesByCategory prints all generators in an industry
func PrintTypesByCategory(industry string) error {
	engine := generators.GetGlobalEngine()
	generators, err := engine.ListGenerators(industry)
	if err != nil {
		return fmt.Errorf("industry '%s' not found", industry)
	}

	fmt.Printf("Generators in industry '%s':\n", industry)
	for _, generator := range generators {
		fmt.Printf("  %s\n", generator)
	}
	return nil
}

// PrintAllTypes prints all generators grouped by industry
func PrintAllTypes() {
	engine := generators.GetGlobalEngine()
	allGenerators := engine.GetAllGenerators()

	// Print generators by industry
	for industry, generatorList := range allGenerators {
		fmt.Printf("\nIndustry: %s\n", industry)
		fmt.Println(strings.Repeat("-", len(industry)+10))

		for _, generator := range generatorList {
			fmt.Printf("  %s\n", generator)
		}
	}
}

// PrintTypeDetails prints detailed information about a generator
func PrintTypeDetails(typeName string) {
	fmt.Printf("Generator: %s\n", typeName)
	fmt.Printf("Use --industry flag to specify industry (base, health, aviation)\n")
	fmt.Printf("Use various parameter flags for customization\n")
}
