package commandline

import (
	"fmt"
	"io"

	"github.com/UiPath/uipathcli/plugin"
)

type SimplePluginCommand struct{}

func (c SimplePluginCommand) Command() plugin.Command {
	return *plugin.NewCommand("mypluginservice", "my-plugin-command", "This is a simple plugin command", []plugin.CommandParameter{}, false)
}

func (c SimplePluginCommand) Execute(context plugin.ExecutionContext, output io.Writer) error {
	output.Write([]byte("Simple plugin command output"))
	return nil
}

type ContextPluginCommand struct {
	Context plugin.ExecutionContext
}

func (c ContextPluginCommand) Command() plugin.Command {
	return *plugin.NewCommand("mypluginservice", "my-plugin-command", "This is a simple plugin command", []plugin.CommandParameter{
		*plugin.NewCommandParameter("filter", plugin.ParameterTypeString, "This is a filter", false),
	}, false)
}

func (c *ContextPluginCommand) Execute(context plugin.ExecutionContext, output io.Writer) error {
	c.Context = context
	output.Write([]byte("Success"))
	return nil
}

type ErrorPluginCommand struct{}

func (c ErrorPluginCommand) Command() plugin.Command {
	return *plugin.NewCommand("mypluginservice", "my-failed-command", "This command always fails", []plugin.CommandParameter{}, false)
}

func (c ErrorPluginCommand) Execute(context plugin.ExecutionContext, output io.Writer) error {
	return fmt.Errorf("Internal server error when calling mypluginservice")
}

type HideOperationPluginCommand struct{}

func (c HideOperationPluginCommand) Command() plugin.Command {
	return *plugin.NewCommand("mypluginservice", "my-hidden-command", "This command should not be shown", []plugin.CommandParameter{}, true)
}

func (c HideOperationPluginCommand) Execute(context plugin.ExecutionContext, output io.Writer) error {
	return fmt.Errorf("my-hidden-command is not supported")
}

type ParametrizedPluginCommand struct{}

func (c ParametrizedPluginCommand) Command() plugin.Command {
	return *plugin.NewCommand("mypluginservice", "my-parametrized-command", "This is a plugin command with parameters", []plugin.CommandParameter{
		*plugin.NewCommandParameter("take", plugin.ParameterTypeInteger, "This is a take parameter", true),
	}, false)
}

func (c ParametrizedPluginCommand) Execute(context plugin.ExecutionContext, output io.Writer) error {
	output.Write([]byte("Parametrized plugin command output"))
	return nil
}
