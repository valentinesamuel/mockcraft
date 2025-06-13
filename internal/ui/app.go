package ui

import (
	"fmt"
	"strings"

	"github.com/valentinesamuel/mockcraft/internal/generators/types"
)

// PrintCategories prints all available categories
func PrintCategories() {
	categories := types.GetAllCategories()
	fmt.Println("Available categories:")
	for _, cat := range categories {
		fmt.Printf("  %s\n", cat)
	}
}

// PrintTypesByCategory prints all types in a category
func PrintTypesByCategory(category string) error {
	typeList := types.GetTypesByCategory(category)
	if len(typeList) == 0 {
		return fmt.Errorf("category '%s' not found", category)
	}

	fmt.Printf("Types in category '%s':\n", category)
	for _, t := range typeList {
		fmt.Printf("  %s: %s\n", t.Name, t.Description)
	}
	return nil
}

// PrintAllTypes prints all types grouped by industry
func PrintAllTypes() {
	// Get all types
	allTypes := make(map[string][]types.TypeDefinition)

	// Group types by industry
	for _, t := range types.GetAllTypes() {
		industry := t.Industry
		if industry == "" {
			industry = "base"
		}
		allTypes[industry] = append(allTypes[industry], t)
	}

	// Print types by industry
	for industry, typeList := range allTypes {
		fmt.Printf("\nIndustry: %s\n", industry)
		fmt.Println(strings.Repeat("-", len(industry)+10))

		for _, t := range typeList {
			fmt.Printf("  %s: %s\n", t.Name, t.Description)
		}
	}
}

// PrintTypeDetails prints detailed information about a type
func PrintTypeDetails(typeName string) {
	typeDef := types.GetTypeByName(typeName)
	if typeDef == nil {
		fmt.Printf("Type '%s' not found\n", typeName)
		return
	}

	fmt.Printf("Type: %s\n", typeDef.Name)
	fmt.Printf("Description: %s\n", typeDef.Description)
	fmt.Printf("Example: %s\n", typeDef.Example)
	fmt.Printf("Category: %s\n", typeDef.Category)

	if len(typeDef.Parameters) > 0 {
		fmt.Println("\nParameters:")
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
}
