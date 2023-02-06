package parser

import "github.com/UiPath/uipathcli/plugin"

type Operation struct {
	Name        string
	Description string
	Method      string
	Route       string
	ContentType string
	Parameters  []Parameter
	Plugin      plugin.CommandPlugin
	Hidden      bool
	Category    *OperationCategory
}

func NewOperation(name string, description string, method string, route string, contentType string, parameters []Parameter, plugin plugin.CommandPlugin, hidden bool, category *OperationCategory) *Operation {
	return &Operation{name, description, method, route, contentType, parameters, plugin, hidden, category}
}
