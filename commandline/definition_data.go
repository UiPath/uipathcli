package commandline

type DefinitionData struct {
	Name string
	Data []byte
}

func NewDefinitionData(name string, data []byte) *DefinitionData {
	return &DefinitionData{name, data}
}
