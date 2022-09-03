package executor

type TokenRequest struct {
	BaseUri      string
	ClientId     string
	ClientSecret string
	Insecure     bool
}

func NewTokenRequest(baseUri string, clientId string, clientSecret string, insecure bool) *TokenRequest {
	return &TokenRequest{baseUri, clientId, clientSecret, insecure}
}
