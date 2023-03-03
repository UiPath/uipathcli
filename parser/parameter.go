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
)

func NewParameter(name string, t string, description string, in string, fieldName string, required bool, defaultValue interface{}, allowedValues []interface{}, parameters []Parameter) *Parameter {
	return &Parameter{name, t, description, in, fieldName, required, defaultValue, allowedValues, parameters}
}
