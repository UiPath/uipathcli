package parser

import (
	"net/url"

	"github.com/UiPath/uipathcli/plugin"
)

// The Operation is the definition of a route which can be called by the CLI.
//
// It holds all the information needed to make the call, like
// HTTP method, Route, Parameters, etc...
type Operation struct {
	Name        string
	Summary     string
	Description string
	Method      string
	BaseUri     url.URL
	Route       string
	ContentType string
	Parameters  []Parameter
	Plugin      plugin.CommandPlugin
	Hidden      bool
	Category    *OperationCategory
}

func NewOperation(name string, summary string, description string, method string, baseUri url.URL, route string, contentType string, parameters []Parameter, plugin plugin.CommandPlugin, hidden bool, category *OperationCategory) *Operation {
	return &Operation{name, summary, description, method, baseUri, route, contentType, parameters, plugin, hidden, category}
}
