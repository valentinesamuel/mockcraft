package generators

import (
	"testing"
	"time"
)

// TestUser demonstrates how to use the struct generator
type TestUser struct {
	Name     string    `fake:"{firstname}"`
	Email    string    `fake:"{email}"`
	Age      int       `fake:"{number:18,65}"`
	Birthday time.Time `fake:"{date}"`
	Active   bool      `fake:"{bool}"`
}

// Fake implements the Fakeable interface for TestUser
func (u *TestUser) Fake() error {
	u.Name = "Custom Name"
	u.Email = "custom@email.com"
	u.Age = 25
	u.Birthday = time.Now()
	u.Active = true
	return nil
}

func TestStructGenerator(t *testing.T) {
	// Create a new generator
	gen := NewStructGenerator()

	// Add custom field function
	gen.AddFieldFunc("Name", func() interface{} {
		return "John Doe"
	})

	// Test with a struct that implements Fakeable
	t.Run("Fakeable interface", func(t *testing.T) {
		user := &TestUser{}
		if err := gen.Generate(user); err != nil {
			t.Fatalf("Failed to generate: %v", err)
		}

		// Verify the custom Fake() method was used
		if user.Name != "Custom Name" {
			t.Errorf("Expected Name to be 'Custom Name', got %q", user.Name)
		}
		if user.Email != "custom@email.com" {
			t.Errorf("Expected Email to be 'custom@email.com', got %q", user.Email)
		}
		if user.Age != 25 {
			t.Errorf("Expected Age to be 25, got %d", user.Age)
		}
		if !user.Active {
			t.Error("Expected Active to be true")
		}
	})

	// Test with a struct that doesn't implement Fakeable
	t.Run("Default generation", func(t *testing.T) {
		type SimpleUser struct {
			Name  string
			Email string
			Age   int
		}

		user := &SimpleUser{}
		if err := gen.Generate(user); err != nil {
			t.Fatalf("Failed to generate: %v", err)
		}

		// Verify the custom field function was used for Name
		if user.Name != "John Doe" {
			t.Errorf("Expected Name to be 'John Doe', got %q", user.Name)
		}

		// Verify other fields were generated
		if user.Email == "" {
			t.Error("Expected Email to be generated")
		}
		if user.Age == 0 {
			t.Error("Expected Age to be generated")
		}
	})
}
