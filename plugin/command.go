package plugin

type Command struct {
	Service     string
	Name        string
	Description string
	Parameters  []CommandParameter
	Hidden      bool
	Category    *CommandCategory
}

func (c *Command) WithCategory(name string, description string) *Command {
	c.Category = NewCommandCategory(name, description)
	return c
}

func NewCommand(service string, name string, description string, parameters []CommandParameter, hidden bool) *Command {
	return &Command{service, name, description, parameters, hidden, nil}
}
