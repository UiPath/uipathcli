package parser

// The Definition provides the high-level information about all operations of the service
type Definition struct {
	Name        string
	Description string
	Operations  []Operation
}

func NewDefinition(name string, description string, operations []Operation) *Definition {
	return &Definition{name, description, operations}
}
