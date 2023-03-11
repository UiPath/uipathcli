package parser

import "testing"

func TestConvertsToSnakeCase(t *testing.T) {
	t.Run("CamelCase", func(t *testing.T) { ConvertsToSnakeCase(t, "myValue", "my-value") })
	t.Run("PascalCase", func(t *testing.T) { ConvertsToSnakeCase(t, "MyValue", "my-value") })
	t.Run("ExistingDash", func(t *testing.T) { ConvertsToSnakeCase(t, "My-Value", "my-value") })
}
func ConvertsToSnakeCase(t *testing.T, input string, expected string) {
	result := ToSnakeCase(input)
	if result != expected {
		t.Errorf("Expected '%v' to be converted to snake case, but got: %v", expected, result)
	}
}
