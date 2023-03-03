package executor

// The ExecutionContextParameters contains list of parameters which are used
// by the executor to dynamically build the request.
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
