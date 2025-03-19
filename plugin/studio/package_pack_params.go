package studio

import (
	"net/url"

	"github.com/UiPath/uipathcli/auth"
)

type packagePackParams struct {
	Organization   string
	Tenant         string
	BaseUri        url.URL
	AuthToken      *auth.AuthToken
	Source         string
	Destination    string
	PackageVersion string
	AutoVersion    bool
	OutputType     string
	SplitOutput    bool
	ReleaseNotes   string
}

func newPackagePackParams(
	organization string,
	tenant string,
	baseUri url.URL,
	authToken *auth.AuthToken,
	source string,
	destination string,
	packageVersion string,
	autoVersion bool,
	outputType string,
	splitOutput bool,
	releaseNotes string) *packagePackParams {
	return &packagePackParams{organization, tenant, baseUri, authToken, source, destination, packageVersion, autoVersion, outputType, splitOutput, releaseNotes}
}
