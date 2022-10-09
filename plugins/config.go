package plugins

type Config struct {
	Authenticators []AuthenticatorConfig
}

type AuthenticatorConfig struct {
	Name string
	Path string
}
