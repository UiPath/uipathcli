package executor

import (
	"net/url"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/plugin"
)

type ExecutionContext struct {
	Organization string
	Tenant       string
	Method       string
	BaseUri      url.URL
	Route        string
	ContentType  string
	Input        *FileReference
	Parameters   ExecutionContextParameters
	AuthConfig   config.AuthConfig
	Insecure     bool
	Debug        bool
	Plugin       plugin.CommandPlugin
}

func NewExecutionContext(
	organization string,
	tenant string,
	method string,
	uri url.URL,
	route string,
	contentType string,
	input *FileReference,
	parameters ExecutionContextParameters,
	authConfig config.AuthConfig,
	insecure bool,
	debug bool,
	plugin plugin.CommandPlugin) *ExecutionContext {
	return &ExecutionContext{organization, tenant, method, uri, route, contentType, input, parameters, authConfig, insecure, debug, plugin}
}
