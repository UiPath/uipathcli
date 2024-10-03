package parser

// OperationCategory allows grouping multiple operations under a common resource.
type OperationCategory struct {
	Name        string
	Summary     string
	Description string
}

func NewOperationCategory(name string, summary string, description string) *OperationCategory {
	return &OperationCategory{name, summary, description}
}
