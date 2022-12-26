package plugin

type Command struct {
	Service     string
	Name        string
	Description string
	Parameters  []CommandParameter
	Hidden      bool
}

func NewCommand(service string, name string, description string, parameters []CommandParameter, hidden bool) *Command {
	return &Command{service, name, description, parameters, hidden}
}
