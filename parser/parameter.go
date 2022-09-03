package parser

type Parameter struct {
	Name        string
	Type        string
	Description string
	In          string
	FieldName   string
	Required    bool
	Parameters  []Parameter
}

const (
	ParameterTypeString  = "string"
	ParameterTypeInteger = "integer"
	ParameterTypeNumber  = "number"
	ParameterTypeBoolean = "boolean"
)

const (
	ParameterInPath   = "path"
	ParameterInQuery  = "query"
	ParameterInHeader = "header"
	ParameterInBody   = "body"
	ParameterInForm   = "form"
)

func NewParameter(name string, t string, description string, in string, fieldName string, required bool, parameters []Parameter) *Parameter {
	return &Parameter{name, t, description, in, fieldName, required, parameters}
}
