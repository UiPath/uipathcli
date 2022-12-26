package plugin

type CommandPlugin interface {
	Command() Command
	Execute(context ExecutionContext) (string, error)
}
