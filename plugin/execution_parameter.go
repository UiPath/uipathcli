package plugin

type ExecutionParameter struct {
	Name  string
	Value interface{}
}

func NewExecutionParameter(name string, value interface{}) *ExecutionParameter {
	return &ExecutionParameter{name, value}
}
