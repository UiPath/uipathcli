package api

type TokenRequest struct {
	GrantType    string
	Scopes       string
	ClientId     string
	ClientSecret string
	Code         string
	CodeVerifier string
	RedirectUri  string
	Properties   map[string]string
}

func NewTokenRequest(grantType string, scopes string, clientId string, clientSecret string, properties map[string]string) *TokenRequest {
	return &TokenRequest{grantType, scopes, clientId, clientSecret, "", "", "", properties}
}

func NewAuthorizationCodeTokenRequest(clientId string, code string, codeVerifier string, redirectUrl string) *TokenRequest {
	return &TokenRequest{"authorization_code", "", clientId, "", code, codeVerifier, redirectUrl, map[string]string{}}
}
