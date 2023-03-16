package config

// PluginConfig keeps metadata about the configured plugins.
//
// Currently supports only external authenticators.
// Example:
// https://github.com/UiPath/uipathcli-authenticator-k8s
type PluginConfig struct {
	Authenticators []AuthenticatorPluginConfig
}

// AuthenticatorPluginConfig holds the information about how to execute the
// external authenticator.
//
// Authenticator plugins require a name and pathto the external executable.
type AuthenticatorPluginConfig struct {
	Name string
	Path string
}
