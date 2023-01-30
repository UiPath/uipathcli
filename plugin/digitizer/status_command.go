package digitzer

import (
	"fmt"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
)

type StatusCommand struct{}

func (c StatusCommand) Command() plugin.Command {
	return *plugin.NewCommand("du-digitizer", "status", "Get Digitization Operation Result", []plugin.CommandParameter{}, true)
}

func (c StatusCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	return fmt.Errorf("Status command not supported")
}
