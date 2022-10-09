package auth

type tokenRequest struct {
	BaseUri      string
	ClientId     string
	ClientSecret string
	Insecure     bool
}

func newTokenRequest(baseUri string, clientId string, clientSecret string, insecure bool) *tokenRequest {
	return &tokenRequest{baseUri, clientId, clientSecret, insecure}
}
