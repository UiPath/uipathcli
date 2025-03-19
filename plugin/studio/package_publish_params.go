package studio

import (
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/network"
)

type packagePublishParams struct {
	Source   string
	Name     string
	Version  string
	BaseUri  string
	Auth     plugin.AuthResult
	Debug    bool
	Settings network.HttpClientSettings
}

func newPackagePublishParams(
	source string,
	name string,
	version string,
	baseUri string,
	auth plugin.AuthResult,
	debug bool,
	settings network.HttpClientSettings) *packagePublishParams {
	return &packagePublishParams{source, name, version, baseUri, auth, debug, settings}
}
