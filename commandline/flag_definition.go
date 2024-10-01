package commandline

type FlagOptionFunc func(*FlagDefinition)

// The FlagDefinition contains the metadata and builder methods for creating
// command line flags.
type FlagDefinition struct {
	Name         string
	Summary      string
	Type         FlagType
	EnvVarName   string
	DefaultValue interface{}
	Hidden       bool
	Required     bool
}

func (f *FlagDefinition) WithDefaultValue(value interface{}) *FlagDefinition {
	f.DefaultValue = value
	return f
}

func (f *FlagDefinition) WithEnvVarName(envVarName string) *FlagDefinition {
	f.EnvVarName = envVarName
	return f
}

func (f *FlagDefinition) WithHidden(hidden bool) *FlagDefinition {
	f.Hidden = hidden
	return f
}

func (f *FlagDefinition) WithRequired(required bool) *FlagDefinition {
	f.Required = required
	return f
}

func NewFlag(name string, summary string, flagType FlagType) *FlagDefinition {
	return &FlagDefinition{
		name,
		summary,
		flagType,
		"",
		nil,
		false,
		false,
	}
}
