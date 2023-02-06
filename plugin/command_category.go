package plugin

type CommandCategory struct {
	Name        string
	Description string
}

func NewCommandCategory(name string, description string) *CommandCategory {
	return &CommandCategory{name, description}
}
