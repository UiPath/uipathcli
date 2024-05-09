package auth

import "net/url"

type bearerAuthenticatorConfig struct {
	GrantType    string
	Scopes       string
	ClientId     string
	ClientSecret string
	Properties   map[string]string
	IdentityUri  url.URL
}

func newBearerAuthenticatorConfig(
	grantType string,
	scopes string,
	clientId string,
	clientSecret string,
	properties map[string]string,
	identityUri url.URL) *bearerAuthenticatorConfig {
	return &bearerAuthenticatorConfig{grantType, scopes, clientId, clientSecret, properties, identityUri}
}
