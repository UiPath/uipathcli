package commandline

// DefinitionData contains the name of the definition file and its data.
type DefinitionData struct {
	Name           string
	ServiceVersion string
	Data           []byte
}

func NewDefinitionData(name string, serviceVersion string, data []byte) *DefinitionData {
	return &DefinitionData{name, serviceVersion, data}
}
