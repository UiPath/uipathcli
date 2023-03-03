package config

import (
	"gopkg.in/yaml.v2"
)

// PluginConfigProvider parses the plugin configuration file
type PluginConfigProvider struct {
	PluginConfigStore PluginConfigStore
	pluginConfig      PluginConfig
}

func (cp *PluginConfigProvider) Load() error {
	data, err := cp.PluginConfigStore.Read()
	if err != nil {
		return err
	}
	var pluginsYaml pluginsYaml
	err = yaml.Unmarshal(data, &pluginsYaml)
	if err != nil {
		return err
	}
	cp.pluginConfig = cp.convertToConfig(pluginsYaml)
	return nil
}

func (cp PluginConfigProvider) Config() PluginConfig {
	return cp.pluginConfig
}

func (cp PluginConfigProvider) convertToConfig(plugins pluginsYaml) PluginConfig {
	authenticators := []AuthenticatorPluginConfig{}
	for _, authenticator := range plugins.Authenticators {
		authenticators = append(authenticators, AuthenticatorPluginConfig(authenticator))
	}
	return PluginConfig{
		Authenticators: authenticators,
	}
}
