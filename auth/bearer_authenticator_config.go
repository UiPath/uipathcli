package auth

import "net/url"

type BearerAuthenticatorConfig struct {
	GrantType    string
	Scopes       string
	ClientId     string
	ClientSecret string
	Properties   map[string]string
	IdentityUri  *url.URL
}

func NewBearerAuthenticatorConfig(
	grantType string,
	scopes string,
	clientId string,
	clientSecret string,
	properties map[string]string,
	identityUri *url.URL) *BearerAuthenticatorConfig {
	return &BearerAuthenticatorConfig{grantType, scopes, clientId, clientSecret, properties, identityUri}
}
