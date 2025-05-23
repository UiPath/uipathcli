package parser

// Parameter contains all the information about a parameter for an operation.
type Parameter struct {
	Name          string
	Type          string
	Description   string
	In            string
	FieldName     string
	Required      bool
	DefaultValue  interface{}
	AllowedValues []interface{}
	Hidden        bool
	Parameters    []Parameter
}

const (
	ParameterTypeString       = "string"
	ParameterTypeBinary       = "binary"
	ParameterTypeInteger      = "integer"
	ParameterTypeNumber       = "number"
	ParameterTypeBoolean      = "boolean"
	ParameterTypeObject       = "object"
	ParameterTypeStringArray  = "stringArray"
	ParameterTypeIntegerArray = "integerArray"
	ParameterTypeNumberArray  = "numberArray"
	ParameterTypeBooleanArray = "booleanArray"
	ParameterTypeObjectArray  = "objectArray"
)

const (
	ParameterInPath   = "path"
	ParameterInQuery  = "query"
	ParameterInHeader = "header"
	ParameterInBody   = "body"
	ParameterInForm   = "form"
	ParameterInCustom = "custom"
)

func (p Parameter) IsArray() bool {
	return p.Type == ParameterTypeBooleanArray ||
		p.Type == ParameterTypeIntegerArray ||
		p.Type == ParameterTypeNumberArray ||
		p.Type == ParameterTypeObjectArray ||
		p.Type == ParameterTypeStringArray
}

func NewParameter(name string, t string, description string, in string, fieldName string, required bool, defaultValue interface{}, allowedValues []interface{}, hidden bool, parameters []Parameter) *Parameter {
	return &Parameter{name, t, description, in, fieldName, required, defaultValue, allowedValues, hidden, parameters}
}
