package auth

import "time"

type tokenResponse struct {
	AccessToken  string
	ExpiresAt    time.Time
	RefreshToken *string
}

func newTokenResponse(accessToken string, expiresAt time.Time, refreshToken *string) *tokenResponse {
	return &tokenResponse{accessToken, expiresAt, refreshToken}
}
