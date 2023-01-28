package executor

import (
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
)

type Executor interface {
	Call(context ExecutionContext, writer output.OutputWriter, logger log.Logger) error
}
