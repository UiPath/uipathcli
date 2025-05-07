package config

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

const DefaultProfile = "default"

// ConfigProvider parses the config file with the profiles.
type ConfigProvider struct {
	store    ConfigStore
	profiles []profileYaml
}

func (p *ConfigProvider) Load() error {
	data, err := p.store.Read()
	if err != nil {
		return err
	}
	var config profilesYaml
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("Error parsing configuration file: %w", err)
	}
	p.profiles = config.Profiles
	return nil
}

func (p *ConfigProvider) Update(profileName string, config Config) error {
	profile := profileYaml{
		Name: profileName,
	}
	index := -1
	for i, p := range p.profiles {
		if p.Name == profileName {
			index = i
			profile = p
		}
	}
	profile.Uri = urlYaml{config.Uri}
	profile.Insecure = config.Insecure
	profile.Debug = config.Debug
	profile.Organization = config.Organization
	profile.Tenant = config.Tenant
	profile.Auth = config.Auth.Config
	profile.Header = config.Header
	profile.Parameter = config.Parameter
	profile.ServiceVersion = config.ServiceVersion

	if index == -1 {
		p.profiles = append(p.profiles, profile)
	} else {
		p.profiles[index] = profile
	}

	data, err := yaml.Marshal(profilesYaml{Profiles: p.profiles})
	if err != nil {
		return fmt.Errorf("Error updating configuration: %w", err)
	}
	return p.store.Write(data)
}

func (p *ConfigProvider) convertToConfig(profile profileYaml) Config {
	if profile.Auth == nil {
		profile.Auth = map[string]interface{}{}
	}
	if profile.Parameter == nil {
		profile.Parameter = map[string]string{}
	}
	if profile.Header == nil {
		profile.Header = map[string]string{}
	}
	return Config{
		Organization: profile.Organization,
		Tenant:       profile.Tenant,
		Uri:          profile.Uri.URL,
		Parameter:    profile.Parameter,
		Header:       profile.Header,
		Auth: AuthConfig{
			Type:   fmt.Sprint(profile.Auth["type"]),
			Config: profile.Auth,
		},
		Insecure:       profile.Insecure,
		Debug:          profile.Debug,
		Output:         profile.Output,
		ServiceVersion: profile.ServiceVersion,
	}
}

func (p *ConfigProvider) New() Config {
	profile := profileYaml{}
	return p.convertToConfig(profile)
}

func (p *ConfigProvider) Config(name string) *Config {
	for _, profile := range p.profiles {
		if profile.Name == name {
			config := p.convertToConfig(profile)
			return &config
		}
	}

	if name == DefaultProfile {
		config := p.New()
		return &config
	}
	return nil
}

func NewConfigProvider(store ConfigStore) *ConfigProvider {
	return &ConfigProvider{
		store: store,
	}
}
