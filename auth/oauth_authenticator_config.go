package auth

import "net/url"

type OAuthAuthenticatorConfig struct {
	ClientId    string
	RedirectUrl url.URL
	Scopes      string
}

func NewOAuthAuthenticatorConfig(
	clientId string,
	redirectUrl url.URL,
	scopes string) *OAuthAuthenticatorConfig {
	return &OAuthAuthenticatorConfig{clientId, redirectUrl, scopes}
}
