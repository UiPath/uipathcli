package pull

import (
	"net/url"

	"github.com/UiPath/uipathcli/plugin"
)

type solutionPullParams struct {
	SolutionId   string
	Destination  string
	BaseUri      url.URL
	Organization string
	Auth         plugin.AuthResult
	Debug        bool
	Settings     plugin.ExecutionSettings
}

func newSolutionPullParams(
	solutionId string,
	destination string,
	baseUri url.URL,
	organization string,
	auth plugin.AuthResult,
	debug bool,
	settings plugin.ExecutionSettings) *solutionPullParams {
	return &solutionPullParams{solutionId, destination, baseUri, organization, auth, debug, settings}
}
