package parser

type OperationCategory struct {
	Name        string
	Description string
}

func NewOperationCategory(name string, description string) *OperationCategory {
	return &OperationCategory{name, description}
}
