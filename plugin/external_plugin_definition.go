package plugin

type ExternalPluginDefinition struct {
	Url         string
	ArchiveName string
	ArchiveType ArchiveType
	Executable  string
}

func NewExternalPluginDefinition(url string, archiveName string, archiveType ArchiveType, executable string) *ExternalPluginDefinition {
	return &ExternalPluginDefinition{url, archiveName, archiveType, executable}
}
