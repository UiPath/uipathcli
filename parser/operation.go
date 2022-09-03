package parser

type Operation struct {
	Name        string
	Description string
	Method      string
	Route       string
	Parameters  []Parameter
}

func NewOperation(name string, description string, method string, route string, parameters []Parameter) *Operation {
	return &Operation{name, description, method, route, parameters}
}
