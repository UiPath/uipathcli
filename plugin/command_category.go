package plugin

// CommandCategory allows grouping multiple operations under a common resource.
//
// Example command with category:
// uipath service category operation --parameter my-value
type CommandCategory struct {
	Name        string
	Summary     string
	Description string
}

func NewCategory(name string, summary string, description string) *CommandCategory {
	return &CommandCategory{name, summary, description}
}
