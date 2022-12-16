package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const DefaultProfile = "default"
const configFilePermissions = 0600

type ConfigProvider struct {
	profiles       []profileYaml
	ConfigFileName string
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

func (cp *ConfigProvider) Update(profileName string, clientId string, clientSecret string, organization string, tenant string) error {
	profile := profileYaml{
		Name: profileName,
		Auth: map[string]interface{}{},
		Path: map[string]string{},
	}
	newProfile := true
	for _, p := range cp.profiles {
		if p.Name == profileName {
			profile = p
			newProfile = false
		}
	}
	if clientId != "" {
		profile.Auth["clientId"] = clientId
	}
	if clientSecret != "" {
		profile.Auth["clientSecret"] = clientSecret
	}
	if organization != "" {
		profile.Path["organization"] = organization
	}
	if tenant != "" {
		profile.Path["tenant"] = tenant
	}
	if newProfile {
		cp.profiles = append(cp.profiles, profile)
	}

	data, err := yaml.Marshal(profilesYaml{Profiles: cp.profiles})
	if err != nil {
		return fmt.Errorf("Error updating configuration: %v", err)
	}
	err = os.MkdirAll(filepath.Dir(cp.ConfigFileName), configFilePermissions)
	if err != nil {
		return fmt.Errorf("Error creating configuration folder: %v", err)
	}
	err = os.WriteFile(cp.ConfigFileName, data, configFilePermissions)
	if err != nil {
		return fmt.Errorf("Error updating configuration file: %v", err)
	}
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
