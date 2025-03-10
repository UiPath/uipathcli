package auth

// AuthenticatorRequest describes the request which needs to be authenticated.
type AuthenticatorRequest struct {
	URL    string
	Header map[string]string
}

func NewAuthenticatorRequest(
	url string,
	header map[string]string) *AuthenticatorRequest {
	return &AuthenticatorRequest{url, header}
}
