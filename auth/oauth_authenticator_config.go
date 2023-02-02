package auth

import "net/url"

type OAuthAuthenticatorConfig struct {
	ClientId    string
	RedirectUrl url.URL
	Scopes      string
	IdentityUri *url.URL
}

func NewOAuthAuthenticatorConfig(
	clientId string,
	redirectUrl url.URL,
	scopes string,
	identityUri *url.URL) *OAuthAuthenticatorConfig {
	return &OAuthAuthenticatorConfig{clientId, redirectUrl, scopes, identityUri}
}
