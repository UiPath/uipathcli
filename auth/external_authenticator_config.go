package auth

// ExternalAuthenticatorConfig keeps the configuration values for the external authenticator.
type ExternalAuthenticatorConfig struct {
	Name string
	Path string
}

func NewExternalAuthenticatorConfig(
	name string,
	path string) *ExternalAuthenticatorConfig {
	return &ExternalAuthenticatorConfig{name, path}
}
