package parser

import "github.com/UiPath/uipathcli/plugin"

type Operation struct {
	Name        string
	Description string
	Method      string
	Route       string
	Parameters  []Parameter
	Plugin      plugin.CommandPlugin
	Hidden      bool
}

func NewOperation(name string, description string, method string, route string, parameters []Parameter, plugin plugin.CommandPlugin, hidden bool) *Operation {
	return &Operation{name, description, method, route, parameters, plugin, hidden}
}
