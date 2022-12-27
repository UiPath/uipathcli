package executor

import "io"

type Executor interface {
	Call(context ExecutionContext, output io.Writer) error
}
