package executor

import (
	"net/url"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/stream"
)

// The ExecutionContext provides all the data needed by the executor to construct the HTTP
// request including URL, headers and body.
type ExecutionContext struct {
	Organization string
	Tenant       string
	Method       string
	BaseUri      url.URL
	Route        string
	ContentType  string
	Input        stream.Stream
	Parameters   ExecutionParameters
	AuthConfig   config.AuthConfig
	IdentityUri  url.URL
	Plugin       plugin.CommandPlugin
	Debug        bool
	Settings     network.HttpClientSettings
}

func NewExecutionContext(
	organization string,
	tenant string,
	method string,
	uri url.URL,
	route string,
	contentType string,
	input stream.Stream,
	parameters []ExecutionParameter,
	authConfig config.AuthConfig,
	identityUri url.URL,
	plugin plugin.CommandPlugin,
	debug bool,
	settings network.HttpClientSettings) *ExecutionContext {
	return &ExecutionContext{
		organization,
		tenant,
		method,
		uri,
		route,
		contentType,
		input,
		parameters,
		authConfig,
		identityUri,
		plugin,
		debug,
		settings,
	}
}
