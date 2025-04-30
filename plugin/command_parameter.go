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
	Name          string
	Type          string
	Description   string
	Required      bool
	DefaultValue  interface{}
	AllowedValues []interface{}
	Hidden        bool
}

func (p *CommandParameter) WithRequired(required bool) *CommandParameter {
	p.Required = required
	return p
}

func (p *CommandParameter) WithDefaultValue(value interface{}) *CommandParameter {
	p.DefaultValue = value
	return p
}

func (p *CommandParameter) WithAllowedValues(values []interface{}) *CommandParameter {
	p.AllowedValues = values
	return p
}

func (p *CommandParameter) WithHidden(hidden bool) *CommandParameter {
	p.Hidden = hidden
	return p
}

func NewParameter(name string, type_ string, description string) *CommandParameter {
	return &CommandParameter{name, type_, description, false, nil, nil, false}
}
