package config

import (
	"gopkg.in/yaml.v2"
)

type PluginConfigProvider struct{}

func (cp *PluginConfigProvider) Parse(data []byte) (*PluginConfig, error) {
	var pluginsYaml pluginsYaml
	err := yaml.Unmarshal(data, &pluginsYaml)
	if err != nil {
		return nil, err
	}
	config := cp.convertToConfig(pluginsYaml)
	return &config, nil
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
