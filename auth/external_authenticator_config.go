package auth

type ExternalAuthenticatorConfig struct {
	Name string
	Path string
}

func NewExternalAuthenticatorConfig(
	name string,
	path string) *ExternalAuthenticatorConfig {
	return &ExternalAuthenticatorConfig{name, path}
}
