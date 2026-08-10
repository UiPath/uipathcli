package push

import (
	"net/url"

	"github.com/UiPath/uipathcli/plugin"
)

type solutionPushParams struct {
	Source       string
	SolutionId   string
	BaseUri      url.URL
	Organization string
	Auth         plugin.AuthResult
	Debug        bool
	Settings     plugin.ExecutionSettings
}

func newSolutionPushParams(
	source string,
	solutionId string,
	baseUri url.URL,
	organization string,
	auth plugin.AuthResult,
	debug bool,
	settings plugin.ExecutionSettings) *solutionPushParams {
	return &solutionPushParams{source, solutionId, baseUri, organization, auth, debug, settings}
}
