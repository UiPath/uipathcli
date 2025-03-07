package studio

import "github.com/UiPath/uipathcli/plugin"

type packagePublishParams struct {
	Source   string
	Name     string
	Version  string
	BaseUri  string
	Auth     plugin.AuthResult
	Insecure bool
	Debug    bool
}

func newPackagePublishParams(
	source string,
	name string,
	version string,
	baseUri string,
	auth plugin.AuthResult,
	insecure bool,
	debug bool) *packagePublishParams {
	return &packagePublishParams{source, name, version, baseUri, auth, insecure, debug}
}
