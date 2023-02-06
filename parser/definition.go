package parser

type Definition struct {
	Name        string
	Description string
	Operations  []Operation
}

func NewDefinition(name string, description string, operations []Operation) *Definition {
	return &Definition{name, description, operations}
}
