package config

type PluginConfig struct {
	Authenticators []AuthenticatorPluginConfig
}

type AuthenticatorPluginConfig struct {
	Name string
	Path string
}
