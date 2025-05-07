package publish

import (
	"net/url"

	"github.com/UiPath/uipathcli/plugin"
)

type packagePublishParams struct {
	Source       string
	Folder       string
	Name         string
	Description  string
	Version      string
	BaseUri      url.URL
	Organization string
	Tenant       string
	Auth         plugin.AuthResult
	Debug        bool
	Settings     plugin.ExecutionSettings
}

func newPackagePublishParams(
	source string,
	folder string,
	name string,
	description string,
	version string,
	baseUri url.URL,
	organization string,
	tenant string,
	auth plugin.AuthResult,
	debug bool,
	settings plugin.ExecutionSettings) *packagePublishParams {
	return &packagePublishParams{
		source,
		folder,
		name,
		description,
		version,
		baseUri,
		organization,
		tenant,
		auth,
		debug,
		settings,
	}
}
