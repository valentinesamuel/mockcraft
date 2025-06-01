package ui

import (
	"fmt"
	"strings"

	"github.com/valentinesamuel/mockcraft/internal/generators/types"
)

// PrintCategories prints all available categories
func PrintCategories() {
	for _, cat := range types.GetAllCategories() {
		fmt.Printf("\n%s\n", strings.ToUpper(cat.Name))
		fmt.Printf("  %s\n", cat.Description)
		fmt.Printf("  Types: %d\n", len(cat.Types))
	}
}

// PrintTypesByCategory prints all types in a category
func PrintTypesByCategory(category string) error {
	types := types.GetTypesByCategory(category)
	if types == nil {
		return fmt.Errorf("category '%s' not found", category)
	}

	for _, t := range types {
		fmt.Printf("\n%s\n", t.Name)
		fmt.Printf("  %s\n", t.Description)
		fmt.Printf("  Example: %s\n", t.Example)
		if len(t.Parameters) > 0 {
			fmt.Println("  Parameters:")
			for _, p := range t.Parameters {
				required := ""
				if p.Required {
					required = " (required)"
				}
				fmt.Printf("    --%s%s: %s\n", p.Name, required, p.Description)
				if p.Default != nil {
					fmt.Printf("      Default: %v\n", p.Default)
				}
			}
		}
	}
	return nil
}

// PrintAllTypes prints all types grouped by category
func PrintAllTypes() {
	for _, cat := range types.GetAllCategories() {
		fmt.Printf("\n%s\n", strings.ToUpper(cat.Name))
		fmt.Printf("  %s\n", cat.Description)
		for _, t := range cat.Types {
			fmt.Printf("  - %s: %s\n", t.Name, t.Description)
		}
	}
}
