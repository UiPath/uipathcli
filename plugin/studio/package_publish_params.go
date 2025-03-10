package studio

import "github.com/UiPath/uipathcli/plugin"

type packagePublishParams struct {
	Source   string
	Name     string
	Version  string
	BaseUri  string
	Auth     plugin.AuthResult
	Debug    bool
	Settings plugin.ExecutionSettings
}

func newPackagePublishParams(
	source string,
	name string,
	version string,
	baseUri string,
	auth plugin.AuthResult,
	debug bool,
	settings plugin.ExecutionSettings) *packagePublishParams {
	return &packagePublishParams{source, name, version, baseUri, auth, debug, settings}
}
