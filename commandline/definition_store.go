package commandline

// DefinitionStore is used to provide the names and content of definition files.
type DefinitionStore interface {
	Names() ([]string, error)
	Read(name string) (*DefinitionData, error)
}
