package auth

type TokenAuthenticatorConfig struct {
	Token     string
}

func NewTokenAuthenticatorConfig(
	token string) *TokenAuthenticatorConfig {
	return &TokenAuthenticatorConfig{token}
}
