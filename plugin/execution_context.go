package plugin

import (
	"net/url"

	"github.com/UiPath/uipathcli/utils/stream"
)

// The ExecutionContext provides all the data needed by the plugin to perform the operation.
type ExecutionContext struct {
	Organization string
	Tenant       string
	BaseUri      url.URL
	Auth         AuthResult
	Input        stream.Stream
	Parameters   []ExecutionParameter
	Debug        bool
	Settings     ExecutionSettings
}

func NewExecutionContext(
	organization string,
	tenant string,
	baseUri url.URL,
	auth AuthResult,
	input stream.Stream,
	parameters []ExecutionParameter,
	debug bool,
	settings ExecutionSettings) *ExecutionContext {
	return &ExecutionContext{organization, tenant, baseUri, auth, input, parameters, debug, settings}
}
