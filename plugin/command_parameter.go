package plugin

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

// CommandParameter defines the parameters the plugin command supports.
type CommandParameter struct {
	Name        string
	Type        string
	Description string
	Required    bool
}

func NewCommandParameter(name string, type_ string, description string, required bool) *CommandParameter {
	return &CommandParameter{name, type_, description, required}
}
