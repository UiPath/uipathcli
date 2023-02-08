package plugin

import "net/url"

type ExecutionContext struct {
	BaseUri    url.URL
	Auth       AuthResult
	Input      *FileParameter
	Parameters []ExecutionParameter
	Insecure   bool
	Debug      bool
}

func NewExecutionContext(
	baseUri url.URL,
	auth AuthResult,
	input *FileParameter,
	parameters []ExecutionParameter,
	insecure bool,
	debug bool) *ExecutionContext {
	return &ExecutionContext{baseUri, auth, input, parameters, insecure, debug}
}
