package plugin

import "net/url"

type ExecutionContext struct {
	Organization string
	Tenant       string
	BaseUri      url.URL
	Auth         AuthResult
	Input        *FileParameter
	Parameters   []ExecutionParameter
	Insecure     bool
	Debug        bool
}

func NewExecutionContext(
	organization string,
	tenant string,
	baseUri url.URL,
	auth AuthResult,
	input *FileParameter,
	parameters []ExecutionParameter,
	insecure bool,
	debug bool) *ExecutionContext {
	return &ExecutionContext{organization, tenant, baseUri, auth, input, parameters, insecure, debug}
}
