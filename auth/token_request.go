package auth

type tokenRequest struct {
	BaseUri      string
	GrantType    string
	ClientId     string
	ClientSecret string
	Code         string
	CodeVerifier string
	RedirectUri  string
	Insecure     bool
}

func newClientCredentialTokenRequest(baseUri string, clientId string, clientSecret string, insecure bool) *tokenRequest {
	return &tokenRequest{baseUri, "client_credentials", clientId, clientSecret, "", "", "", insecure}
}

func newAuthorizationCodeTokenRequest(baseUri string, clientId string, code string, codeVerifier string, redirectUrl string, insecure bool) *tokenRequest {
	return &tokenRequest{baseUri, "authorization_code", clientId, "", code, codeVerifier, redirectUrl, insecure}
}
