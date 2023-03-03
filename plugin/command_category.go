package plugin

// CommandCategory allows grouping multiple operations under a common resource.
//
// Example command with category:
// uipath service category operation --parameter my-value
type CommandCategory struct {
	Name        string
	Description string
}

func NewCommandCategory(name string, description string) *CommandCategory {
	return &CommandCategory{name, description}
}
