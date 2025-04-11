package studio

import (
	"net/url"

	"github.com/UiPath/uipathcli/auth"
)

type packageRestoreParams struct {
	Organization string
	Tenant       string
	BaseUri      url.URL
	AuthToken    *auth.AuthToken
	Source       string
	Destination  string
}

func newPackageRestoreParams(
	organization string,
	tenant string,
	baseUri url.URL,
	authToken *auth.AuthToken,
	source string,
	destination string) *packageRestoreParams {
	return &packageRestoreParams{organization, tenant, baseUri, authToken, source, destination}
}
