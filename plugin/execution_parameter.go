package plugin

// An ExecutionParameter is a value which is used by the executor to build the request.
// Parameter values are typicall provided by multiple sources like config files,
// command line arguments and environment variables.
type ExecutionParameter struct {
	Name  string
	Value interface{}
}

func NewExecutionParameter(name string, value interface{}) *ExecutionParameter {
	return &ExecutionParameter{name, value}
}
