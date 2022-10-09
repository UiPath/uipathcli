package auth

type identityResponse struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   float32 `json:"expires_in"`
}
