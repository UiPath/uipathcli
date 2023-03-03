package commandline

// DefinitionData contains the name of the definition file and its data.
type DefinitionData struct {
	Name string
	Data []byte
}

func NewDefinitionData(name string, data []byte) *DefinitionData {
	return &DefinitionData{name, data}
}
