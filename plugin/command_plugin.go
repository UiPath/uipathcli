package plugin

import "io"

type CommandPlugin interface {
	Command() Command
	Execute(context ExecutionContext, output io.Writer) error
}
