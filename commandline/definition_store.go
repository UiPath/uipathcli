package commandline

// DefinitionStore is used to provide the names and content of definition files.
type DefinitionStore interface {
	Names(serviceVersion string) ([]string, error)
	Read(name string, serviceVersion string) (*DefinitionData, error)
}
