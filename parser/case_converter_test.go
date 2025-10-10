package parser

import "testing"

func TestConvertsToSnakeCase(t *testing.T) {
	t.Run("CamelCase", func(t *testing.T) { ConvertsToSnakeCase(t, "myValue", "my-value") })
	t.Run("PascalCase", func(t *testing.T) { ConvertsToSnakeCase(t, "MyValue", "my-value") })
	t.Run("ExistingDash", func(t *testing.T) { ConvertsToSnakeCase(t, "My-Value", "my-value") })
	t.Run("IgnoresDuplicateDash", func(t *testing.T) { ConvertsToSnakeCase(t, "get--Ping", "get-ping") })
	t.Run("IgnoresDuplicateUnderscore", func(t *testing.T) { ConvertsToSnakeCase(t, "get__Ping", "get-ping") })
	t.Run("IgnoresDuplicateSlash", func(t *testing.T) { ConvertsToSnakeCase(t, "get//Ping", "get-ping") })
	t.Run("IgnoresCurlyBrackets", func(t *testing.T) {
		ConvertsToSnakeCase(t, "get/MyResource/{resourceId}", "get-my-resource-resource-id")
	})
	t.Run("IgnoresDollarSign", func(t *testing.T) {
		ConvertsToSnakeCase(t, "get/$MyResource", "get-my-resource")
	})
}
func ConvertsToSnakeCase(t *testing.T, input string, expected string) {
	result := toSnakeCase(input)
	if result != expected {
		t.Errorf("Expected %v, but got: %v", expected, result)
	}
}
