package auth

import (
	"net/url"

	"github.com/UiPath/uipathcli/log"
)

// AuthenticatorContext provides information required for authenticating requests.
type AuthenticatorContext struct {
	Config      map[string]interface{}
	IdentityUri url.URL
	OperationId string
	Insecure    bool
	Debug       bool
	Request     AuthenticatorRequest
	Logger      log.Logger
}

func NewAuthenticatorContext(
	config map[string]interface{},
	identityUri url.URL,
	operationId string,
	insecure bool,
	debug bool,
	request AuthenticatorRequest,
	logger log.Logger,
) *AuthenticatorContext {
	return &AuthenticatorContext{
		config,
		identityUri,
		operationId,
		insecure,
		debug,
		request,
		logger,
	}
}
