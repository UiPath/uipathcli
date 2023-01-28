package plugin

import "net/url"

type ExecutionContext struct {
	BaseUri    url.URL
	Auth       AuthResult
	Parameters []ExecutionParameter
	Insecure   bool
}

func NewExecutionContext(
	baseUri url.URL,
	auth AuthResult,
	parameters []ExecutionParameter,
	insecure bool) *ExecutionContext {
	return &ExecutionContext{baseUri, auth, parameters, insecure}
}
