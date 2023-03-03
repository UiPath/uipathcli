package auth

// AuthenticatorRequest describes the request which needs to be authenticated.
type AuthenticatorRequest struct {
	URL    string            `json:"url"`
	Header map[string]string `json:"header"`
}

func NewAuthenticatorRequest(
	url string,
	header map[string]string) *AuthenticatorRequest {
	return &AuthenticatorRequest{url, header}
}
