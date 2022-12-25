package auth

type tokenResponse struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   float32 `json:"expires_in"`
}

func newTokenResponse(accessToken string, expiresIn float32) *tokenResponse {
	return &tokenResponse{accessToken, expiresIn}
}
