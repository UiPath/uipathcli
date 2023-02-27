package config

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

const DefaultProfile = "default"

type ConfigProvider struct {
	profiles    []profileYaml
	ConfigStore ConfigStore
}

func (cp *ConfigProvider) Load() error {
	data, err := cp.ConfigStore.Read()
	if err != nil {
		return err
	}
	var config profilesYaml
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("Error parsing configuration file: %v", err)
	}
	cp.profiles = config.Profiles
	return nil
}

func (cp *ConfigProvider) Update(profileName string, config Config) error {
	profile := profileYaml{
		Name: profileName,
	}
	index := -1
	for i, p := range cp.profiles {
		if p.Name == profileName {
			index = i
			profile = p
		}
	}
	profile.Organization = config.Organization
	profile.Tenant = config.Tenant
	profile.Auth = config.Auth.Config

	if index == -1 {
		cp.profiles = append(cp.profiles, profile)
	} else {
		cp.profiles[index] = profile
	}

	data, err := yaml.Marshal(profilesYaml{Profiles: cp.profiles})
	if err != nil {
		return fmt.Errorf("Error updating configuration: %v", err)
	}
	return cp.ConfigStore.Write(data)
}

func (cp ConfigProvider) convertToConfig(profile profileYaml) Config {
	if profile.Auth == nil {
		profile.Auth = map[string]interface{}{}
	}
	if profile.Path == nil {
		profile.Path = map[string]string{}
	}
	if profile.Header == nil {
		profile.Header = map[string]string{}
	}
	if profile.Query == nil {
		profile.Query = map[string]string{}
	}
	return Config{
		Organization: profile.Organization,
		Tenant:       profile.Tenant,
		Uri:          profile.Uri.URL,
		Path:         profile.Path,
		Query:        profile.Query,
		Header:       profile.Header,
		Auth: AuthConfig{
			Type:   fmt.Sprintf("%v", profile.Auth["type"]),
			Config: profile.Auth,
		},
		Insecure: profile.Insecure,
		Debug:    profile.Debug,
		Output:   profile.Output,
	}
}

func (cp ConfigProvider) New() Config {
	profile := profileYaml{}
	return cp.convertToConfig(profile)
}

func (cp ConfigProvider) Config(name string) *Config {
	for _, profile := range cp.profiles {
		if profile.Name == name {
			config := cp.convertToConfig(profile)
			return &config
		}
	}

	if name == DefaultProfile {
		config := cp.New()
		return &config
	}
	return nil
}
