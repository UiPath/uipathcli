package auth

import "net/url"

type oauthAuthenticatorConfig struct {
	ClientId     string
	ClientSecret string
	RedirectUrl  url.URL
	Scopes       string
	IdentityUri  url.URL
}

func newOAuthAuthenticatorConfig(
	clientId string,
	clientSecret string,
	redirectUrl url.URL,
	scopes string,
	identityUri url.URL) *oauthAuthenticatorConfig {
	return &oauthAuthenticatorConfig{clientId, clientSecret, redirectUrl, scopes, identityUri}
}
