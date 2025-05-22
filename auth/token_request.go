package auth

import (
	"net/url"

	"github.com/UiPath/uipathcli/utils/network"
)

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
	RefreshToken string
	Settings     network.HttpClientSettings
}

func newTokenRequest(baseUri url.URL, grantType string, scopes string, clientId string, clientSecret string, properties map[string]string, settings network.HttpClientSettings) *tokenRequest {
	return &tokenRequest{baseUri, grantType, scopes, clientId, clientSecret, "", "", "", properties, "", settings}
}

func newAuthorizationCodeTokenRequest(baseUri url.URL, clientId string, clientSecret string, code string, codeVerifier string, redirectUrl string, settings network.HttpClientSettings) *tokenRequest {
	return &tokenRequest{baseUri, "authorization_code", "", clientId, clientSecret, code, codeVerifier, redirectUrl, map[string]string{}, "", settings}
}

func newRefreshTokenRequest(baseUri url.URL, clientId string, clientSecret string, refreshToken string, settings network.HttpClientSettings) *tokenRequest {
	return &tokenRequest{baseUri, "refresh_token", "", clientId, clientSecret, "", "", "", map[string]string{}, refreshToken, settings}
}
