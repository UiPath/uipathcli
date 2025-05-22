package commandline

import (
	"net/url"
)

type configCommandInput struct {
	AuthType    string
	Profile     string
	Debug       bool
	BaseUri     url.URL
	OperationId string
	Insecure    bool
	IdentityUri url.URL
}

func newConfigCommandInput(
	authType string,
	profile string,
	debug bool,
	operationId string,
	insecure bool,
	baseUri url.URL,
	identityUri url.URL,
) *configCommandInput {
	return &configCommandInput{
		AuthType:    authType,
		Profile:     profile,
		Debug:       debug,
		OperationId: operationId,
		Insecure:    insecure,
		BaseUri:     baseUri,
		IdentityUri: identityUri,
	}
}
