package executor

type Executor interface {
	Call(context ExecutionContext) (string, error)
}
