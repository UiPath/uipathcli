package commandline

// DefinitionStore is used to provide the names and content of definition files.
type DefinitionStore interface {
	Names(version string) ([]string, error)
	Read(name string, version string) (*DefinitionData, error)
}
