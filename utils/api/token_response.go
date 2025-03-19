package api

type TokenResponse struct {
	AccessToken string
	ExpiresIn   float32
}

func NewTokenResponse(accessToken string, expiresIn float32) *TokenResponse {
	return &TokenResponse{accessToken, expiresIn}
}
