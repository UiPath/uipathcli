package parser

import (
	"net/url"

	"github.com/UiPath/uipathcli/plugin"
)

type Operation struct {
	Name        string
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

func NewOperation(name string, description string, method string, baseUri url.URL, route string, contentType string, parameters []Parameter, plugin plugin.CommandPlugin, hidden bool, category *OperationCategory) *Operation {
	return &Operation{name, description, method, baseUri, route, contentType, parameters, plugin, hidden, category}
}
