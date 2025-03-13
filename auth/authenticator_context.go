package auth

import "net/url"

// AuthenticatorContext provides information required for authenticating requests.
type AuthenticatorContext struct {
	Type        string
	Config      map[string]interface{}
	IdentityUri url.URL
	OperationId string
	Insecure    bool
	Request     AuthenticatorRequest
}

func NewAuthenticatorContext(
	authType string,
	config map[string]interface{},
	identityUri url.URL,
	operationId string,
	insecure bool,
	request AuthenticatorRequest) *AuthenticatorContext {
	return &AuthenticatorContext{authType, config, identityUri, operationId, insecure, request}
}
