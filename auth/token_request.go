package auth

import "net/url"

type tokenRequest struct {
	BaseUri      url.URL
	GrantType    string
	Scopes       string
	ClientId     string
	ClientSecret string
	Code         string
	CodeVerifier string
	RedirectUri  string
	Properties   map[string]string
	Insecure     bool
}

func newTokenRequest(baseUri url.URL, grantType string, scopes string, clientId string, clientSecret string, properties map[string]string, insecure bool) *tokenRequest {
	return &tokenRequest{baseUri, grantType, scopes, clientId, clientSecret, "", "", "", properties, insecure}
}

func newAuthorizationCodeTokenRequest(baseUri url.URL, clientId string, code string, codeVerifier string, redirectUrl string, insecure bool) *tokenRequest {
	return &tokenRequest{baseUri, "authorization_code", "", clientId, "", code, codeVerifier, redirectUrl, map[string]string{}, insecure}
}
