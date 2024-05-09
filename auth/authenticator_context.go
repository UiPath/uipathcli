package auth

import "net/url"

// AuthenticatorContext provides information required for authenticating requests.
type AuthenticatorContext struct {
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config"`
	IdentityUri url.URL                `json:"identityUri"`
	Debug       bool                   `json:"debug"`
	Insecure    bool                   `json:"insecure"`
	Request     AuthenticatorRequest   `json:"request"`
}

func NewAuthenticatorContext(
	authType string,
	config map[string]interface{},
	identityUri url.URL,
	debug bool,
	insecure bool,
	request AuthenticatorRequest) *AuthenticatorContext {
	return &AuthenticatorContext{authType, config, identityUri, debug, insecure, request}
}
