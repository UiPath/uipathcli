package plugin

import (
	"net/url"

	"github.com/UiPath/uipathcli/utils"
)

// The ExecutionContext provides all the data needed by the plugin to perform the operation.
type ExecutionContext struct {
	Organization string
	Tenant       string
	BaseUri      url.URL
	Auth         AuthResult
	Input        utils.Stream
	Parameters   []ExecutionParameter
	Insecure     bool
	Debug        bool
}

func NewExecutionContext(
	organization string,
	tenant string,
	baseUri url.URL,
	auth AuthResult,
	input utils.Stream,
	parameters []ExecutionParameter,
	insecure bool,
	debug bool) *ExecutionContext {
	return &ExecutionContext{organization, tenant, baseUri, auth, input, parameters, insecure, debug}
}
