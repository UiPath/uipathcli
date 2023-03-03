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
// The Path to the external authenticator executable
type AuthenticatorPluginConfig struct {
	Name string
	Path string
}
