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

// PrintAllTypes prints all types grouped by category
func PrintAllTypes() {
	// Get all categories
	categories := types.GetAllCategories()

	// Print types by category
	for _, cat := range categories {
		fmt.Printf("\nCategory: %s\n", cat)
		fmt.Println(strings.Repeat("-", len(cat)+10))

		typeList := types.GetTypesByCategory(cat)
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
