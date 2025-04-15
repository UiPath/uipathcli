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
	IdentityUri  url.URL
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
	identityUri url.URL,
	input stream.Stream,
	parameters []ExecutionParameter,
	debug bool,
	settings ExecutionSettings,
) *ExecutionContext {
	return &ExecutionContext{
		organization,
		tenant,
		baseUri,
		auth,
		identityUri,
		input,
		parameters,
		debug,
		settings,
	}
}
