package executor

type ExecutionContextParameters struct {
	Path   []ExecutionParameter
	Query  []ExecutionParameter
	Header []ExecutionParameter
	Body   []ExecutionParameter
	Form   []ExecutionParameter
}

func NewExecutionContextParameters(
	path []ExecutionParameter,
	query []ExecutionParameter,
	header []ExecutionParameter,
	body []ExecutionParameter,
	form []ExecutionParameter) *ExecutionContextParameters {
	return &ExecutionContextParameters{path, query, header, body, form}
}
