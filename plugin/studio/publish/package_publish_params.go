package publish

import "github.com/UiPath/uipathcli/plugin"

type packagePublishParams struct {
	Source      string
	Folder      string
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
	folder string,
	name string,
	description string,
	version string,
	baseUri string,
	auth plugin.AuthResult,
	debug bool,
	settings plugin.ExecutionSettings) *packagePublishParams {
	return &packagePublishParams{source, folder, name, description, version, baseUri, auth, debug, settings}
}
