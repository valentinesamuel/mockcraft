package config

import (
	"testing"
)

func TestValidateSchema(t *testing.T) {
	// Test empty schema
	schema := &Schema{}
	if err := ValidateSchema(schema); err == nil {
		t.Error("Expected error for empty schema")
	}

	// Test valid schema
	schema = &Schema{
		Tables: []Table{
			{
				Name:  "users",
				Count: 100,
				Columns: []Column{
					{
						Name:      "id",
						Type:      "uuid",
						Generator: "uuid",
					},
				},
			},
		},
	}
	if err := ValidateSchema(schema); err != nil {
		t.Errorf("Unexpected error for valid schema: %v", err)
	}

	// Test invalid table name
	schema = &Schema{
		Tables: []Table{
			{
				Name:  "",
				Count: 100,
				Columns: []Column{
					{
						Name:      "id",
						Type:      "uuid",
						Generator: "uuid",
					},
				},
			},
		},
	}
	if err := ValidateSchema(schema); err == nil {
		t.Error("Expected error for invalid table name")
	}

	// Test invalid column
	schema = &Schema{
		Tables: []Table{
			{
				Name:  "users",
				Count: 100,
				Columns: []Column{
					{
						Name:      "",
						Type:      "uuid",
						Generator: "uuid",
					},
				},
			},
		},
	}
	if err := ValidateSchema(schema); err == nil {
		t.Error("Expected error for invalid column")
	}

	// Test circular dependency
	schema = &Schema{
		Tables: []Table{
			{
				Name:  "users",
				Count: 100,
				Columns: []Column{
					{
						Name:      "id",
						Type:      "uuid",
						Generator: "uuid",
					},
				},
				Relations: []Relation{
					{
						Type:       "one-to-many",
						FromTable:  "users",
						FromColumn: "id",
						ToTable:    "orders",
						ToColumn:   "user_id",
					},
				},
			},
			{
				Name:  "orders",
				Count: 100,
				Columns: []Column{
					{
						Name:      "id",
						Type:      "uuid",
						Generator: "uuid",
					},
				},
				Relations: []Relation{
					{
						Type:       "one-to-many",
						FromTable:  "orders",
						FromColumn: "id",
						ToTable:    "users",
						ToColumn:   "order_id",
					},
				},
			},
		},
	}
	if err := ValidateSchema(schema); err == nil {
		t.Error("Expected error for circular dependency")
	}
}

func TestValidateRelation(t *testing.T) {
	// Test valid relation
	relation := &Relation{
		Type:       "one-to-many",
		FromTable:  "users",
		FromColumn: "id",
		ToTable:    "orders",
		ToColumn:   "user_id",
	}
	if err := validateRelation(relation); err != nil {
		t.Errorf("Unexpected error for valid relation: %v", err)
	}

	// Test invalid relation type
	relation = &Relation{
		Type:       "invalid",
		FromTable:  "users",
		FromColumn: "id",
		ToTable:    "orders",
		ToColumn:   "user_id",
	}
	if err := validateRelation(relation); err == nil {
		t.Error("Expected error for invalid relation type")
	}

	// Test missing required fields
	relation = &Relation{
		Type:       "one-to-many",
		FromColumn: "id",
		ToTable:    "orders",
		ToColumn:   "user_id",
	}
	if err := validateRelation(relation); err == nil {
		t.Error("Expected error for missing from_table")
	}
}
