package commandline

// DefinitionData contains the name of the definition file and its data.
type DefinitionData struct {
	Name    string
	Version string
	Data    []byte
}

func NewDefinitionData(name string, version string, data []byte) *DefinitionData {
	return &DefinitionData{name, version, data}
}
