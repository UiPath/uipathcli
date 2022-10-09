package auth

type BearerAuthenticatorConfig struct {
	ClientId     string
	ClientSecret string
}

func NewBearerAuthenticatorConfig(
	clientId string,
	clientSecret string) *BearerAuthenticatorConfig {
	return &BearerAuthenticatorConfig{clientId, clientSecret}
}
