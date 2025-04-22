package publish

import "github.com/UiPath/uipathcli/plugin"

type packagePublishParams struct {
	Source      string
	FolderId    int
	Name        string
	Description string
	Version     string
	BaseUri     string
	Auth        plugin.AuthResult
	Debug       bool
	Settings    plugin.ExecutionSettings
}

func newPackagePublishParams(
	source string,
	folderId int,
	name string,
	description string,
	version string,
	baseUri string,
	auth plugin.AuthResult,
	debug bool,
	settings plugin.ExecutionSettings) *packagePublishParams {
	return &packagePublishParams{source, folderId, name, description, version, baseUri, auth, debug, settings}
}
