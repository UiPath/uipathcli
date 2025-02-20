package executor

import (
	"net/url"
	"time"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/plugin"
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
	Insecure     bool
	Timeout      time.Duration
	MaxAttempts  int
	Debug        bool
	IdentityUri  url.URL
	Plugin       plugin.CommandPlugin
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
	insecure bool,
	timeout time.Duration,
	maxAttempts int,
	debug bool,
	identityUri url.URL,
	plugin plugin.CommandPlugin) *ExecutionContext {
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
		insecure,
		timeout,
		maxAttempts,
		debug,
		identityUri,
		plugin,
	}
}
