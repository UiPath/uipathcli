package plugin

import (
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
)

type CommandPlugin interface {
	Command() Command
	Execute(context ExecutionContext, writer output.OutputWriter, logger log.Logger) error
}
