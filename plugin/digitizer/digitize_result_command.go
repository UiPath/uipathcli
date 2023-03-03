package digitzer

import (
	"fmt"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
)

// DigitizeResultCommand only marks the digitize-result command as hidden
//
// The digitize-result command is not needed because the CLI provides a convenient wrapper over
// the async digitizer API to make it seem like it is a single sync call.
type DigitizeResultCommand struct{}

func (c DigitizeResultCommand) Command() plugin.Command {
	return *plugin.NewCommand("du").
		WithCategory("digitization", "Document Digitization").
		WithOperation("digitize-result", "Get Digitization Operation Result").
		IsHidden()
}

func (c DigitizeResultCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	return fmt.Errorf("Digitize result command not supported")
}
