package plugins

import (
	"gopkg.in/yaml.v2"
)

const DefaultProfile = "default"

type ConfigProvider struct{}

func (cp *ConfigProvider) Parse(data []byte) (*Config, error) {
	var pluginsYaml pluginsYaml
	err := yaml.Unmarshal(data, &pluginsYaml)
	if err != nil {
		return nil, err
	}
	config := cp.convertToConfig(pluginsYaml)
	return &config, nil
}

func (cp ConfigProvider) convertToConfig(plugins pluginsYaml) Config {
	authenticators := []AuthenticatorConfig{}
	for _, authenticator := range plugins.Authenticators {
		authenticators = append(authenticators, AuthenticatorConfig(authenticator))
	}
	return Config{
		Authenticators: authenticators,
	}
}
