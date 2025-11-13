package commandline

import (
	"context"

	"github.com/urfave/cli/v3"
)

// The CommandExecContext contains the flag values provided by the user.
type CommandExecContext struct {
	*cli.Command
	Context context.Context
}

// The CommandExecFunc is the function definition for executing a command action.
type CommandExecFunc func(*CommandExecContext) error

// The CommandDefinition contains the metadata and builder methods for creating
// CLI commands.
type CommandDefinition struct {
	Name         string
	Summary      string
	Description  string
	Flags        []*FlagDefinition
	Subcommands  []*CommandDefinition
	HelpTemplate string
	Hidden       bool
	Action       CommandExecFunc
}

func (c *CommandDefinition) WithFlags(flags []*FlagDefinition) *CommandDefinition {
	c.Flags = flags
	return c
}

func (c *CommandDefinition) WithSubcommands(subcommands []*CommandDefinition) *CommandDefinition {
	c.Subcommands = subcommands
	return c
}

func (c *CommandDefinition) WithHidden(hidden bool) *CommandDefinition {
	c.Hidden = hidden
	return c
}

func (c *CommandDefinition) WithHelpTemplate(helpTemplate string) *CommandDefinition {
	c.HelpTemplate = helpTemplate
	return c
}

func (c *CommandDefinition) WithAction(action CommandExecFunc) *CommandDefinition {
	c.Action = action
	return c
}

func NewCommand(name string, summary string, description string) *CommandDefinition {
	return &CommandDefinition{
		name,
		summary,
		description,
		nil,
		nil,
		"",
		false,
		nil,
	}
}
