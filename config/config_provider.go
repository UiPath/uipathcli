package config

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

const DefaultProfile = "default"

type ConfigProvider struct {
	profiles []profileYaml
}

func (cp *ConfigProvider) Load(data []byte) error {
	var config profilesYaml
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("Error parsing configuration file: %v", err)
	}
	cp.profiles = config.Profiles
	return nil
}

func (cp ConfigProvider) convertToConfig(profile profileYaml) Config {
	return Config{
		Uri:    profile.Uri.URL,
		Path:   profile.Path,
		Query:  profile.Query,
		Header: profile.Header,
		Auth: AuthConfig{
			Type:   fmt.Sprintf("%v", profile.Auth["type"]),
			Config: profile.Auth,
		},
		Insecure: profile.Insecure,
		Debug:    profile.Debug,
	}
}

func (cp ConfigProvider) Config(name string) *Config {
	for _, profile := range cp.profiles {
		if profile.Name == name {
			config := cp.convertToConfig(profile)
			return &config
		}
	}

	if name == DefaultProfile {
		return &Config{}
	}
	return nil
}
