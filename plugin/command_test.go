package plugin

import (
	"testing"
)

func TestCommand(t *testing.T) {
	command := NewCommand("my-service").
		WithCategory("my-category", "category description").
		WithOperation("my-operation", "operation description").
		IsHidden()

	if command.Service != "my-service" {
		t.Errorf("Did not return service name, but got: %v", command.Service)
	}
	if command.Category.Name != "my-category" {
		t.Errorf("Did not return category name, but got: %v", command.Category.Name)
	}
	if command.Category.Description != "category description" {
		t.Errorf("Did not return category description, but got: %v", command.Category.Description)
	}
	if command.Name != "my-operation" {
		t.Errorf("Did not return operation name, but got: %v", command.Name)
	}
	if command.Description != "operation description" {
		t.Errorf("Did not return operation description, but got: %v", command.Description)
	}
	if !command.Hidden {
		t.Errorf("Command should be hidden, but it is not")
	}
}
