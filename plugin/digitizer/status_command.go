package plugin_digitizer

import (
	"fmt"

	"github.com/UiPath/uipathcli/plugin"
)

type StatusCommand struct{}

func (c StatusCommand) Command() plugin.Command {
	return *plugin.NewCommand("digitizer", "status", "Get Digitization Operation Result", []plugin.CommandParameter{}, true)
}

func (c StatusCommand) Execute(context plugin.ExecutionContext) (string, error) {
	return "", fmt.Errorf("Status command not supported")
}
