package parser

// OperationCategory allows grouping multiple operations under a common resource.
type OperationCategory struct {
	Name        string
	Description string
}

func NewOperationCategory(name string, description string) *OperationCategory {
	return &OperationCategory{name, description}
}
