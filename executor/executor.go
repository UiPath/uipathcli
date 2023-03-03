// Package executor calls UiPath services.
package executor

import (
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
)

// The Executor interface is an abstraction for carrying out CLI commands.
//
// The ExecutionContext provides all the data needed to execute a command.
// The OutputWriter should be used to output the result of the command.
// The Logger should be used for providing additional information when running a command.
type Executor interface {
	Call(context ExecutionContext, writer output.OutputWriter, logger log.Logger) error
}
