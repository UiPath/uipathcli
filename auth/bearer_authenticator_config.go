package auth

import "net/url"

type BearerAuthenticatorConfig struct {
	ClientId     string
	ClientSecret string
	IdentityUri  *url.URL
}

func NewBearerAuthenticatorConfig(
	clientId string,
	clientSecret string,
	identityUri *url.URL) *BearerAuthenticatorConfig {
	return &BearerAuthenticatorConfig{clientId, clientSecret, identityUri}
}
