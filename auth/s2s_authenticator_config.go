package auth

type S2SAuthenticatorConfig struct {
	ClientId     string
	ClientSecret string
}

func NewS2SAuthenticatorConfig(
	clientId string,
	clientSecret string) *S2SAuthenticatorConfig {
	return &S2SAuthenticatorConfig{clientId, clientSecret}
}
