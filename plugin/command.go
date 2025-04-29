package plugin

// Command is used to define the metadata of the plugin.
//
// Command defines the service name, command name and its available parameters.
type Command struct {
	Service     string
	Name        string
	Summary     string
	Description string
	Parameters  []CommandParameter
	Hidden      bool
	Category    *CommandCategory
}

func (c *Command) WithCategory(name string, summary string, description string) *Command {
	c.Category = NewCategory(name, summary, description)
	return c
}

func (c *Command) WithOperation(name string, summary string, description string) *Command {
	c.Name = name
	c.Summary = summary
	c.Description = description
	return c
}

func (c *Command) WithParameter(parameter *CommandParameter) *Command {
	c.Parameters = append(c.Parameters, *parameter)
	return c
}

func (c *Command) IsHidden() *Command {
	c.Hidden = true
	return c
}

func NewCommand(service string) *Command {
	return &Command{
		Service:    service,
		Parameters: []CommandParameter{},
	}
}
