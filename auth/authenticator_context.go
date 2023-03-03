package auth

// AuthenticatorContext provides information required for authenticating requests.
type AuthenticatorContext struct {
	Type     string                 `json:"type"`
	Config   map[string]interface{} `json:"config"`
	Debug    bool                   `json:"debug"`
	Insecure bool                   `json:"insecure"`
	Request  AuthenticatorRequest   `json:"request"`
}

func NewAuthenticatorContext(
	authType string,
	config map[string]interface{},
	debug bool,
	insecure bool,
	request AuthenticatorRequest) *AuthenticatorContext {
	return &AuthenticatorContext{authType, config, debug, insecure, request}
}
