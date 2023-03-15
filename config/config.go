// Package config parses config and plugin configuration files. It provides APIs to read
// and write config files and manage multiple profiles.
package config

import (
	"fmt"
	"net/url"
)

// The Config structure holds the config data from the selected profile.
type Config struct {
	Uri          *url.URL
	Organization string
	Tenant       string
	Path         map[string]string
	Query        map[string]string
	Header       map[string]string
	Auth         AuthConfig
	Insecure     bool
	Debug        bool
	Output       string
}

// AuthConfig with metadata used for authenticating the caller.
type AuthConfig struct {
	Type   string
	Config map[string]interface{}
}

const clientIdKey = "clientId"
const clientSecretKey = "clientSecret"
const redirectUriKey = "redirectUri"
const scopesKey = "scopes"
const patKey = "pat"

func (c Config) ClientId() string {
	clientId := c.Auth.Config[clientIdKey]
	if clientId == nil {
		return ""
	}
	return fmt.Sprintf("%v", clientId)
}

func (c Config) ClientSecret() string {
	clientSecret := c.Auth.Config[clientSecretKey]
	if clientSecret == nil {
		return ""
	}
	return fmt.Sprintf("%v", clientSecret)
}

func (c Config) RedirectUri() string {
	redirectUri := c.Auth.Config[redirectUriKey]
	if redirectUri == nil {
		return ""
	}
	return fmt.Sprintf("%v", redirectUri)
}

func (c Config) Scopes() string {
	scopes := c.Auth.Config[scopesKey]
	if scopes == nil {
		return ""
	}
	return fmt.Sprintf("%v", scopes)
}

func (c Config) Pat() string {
	pat := c.Auth.Config[patKey]
	if pat == nil {
		return ""
	}
	return fmt.Sprintf("%v", pat)
}

func (c *Config) ConfigureOrgTenant(organization string, tenant string) bool {
	if organization != "" {
		c.Organization = organization
	}
	if tenant != "" {
		c.Tenant = tenant
	}

	return organization != "" || tenant != ""
}

func (c Config) ConfigurePatAuth(pat string) bool {
	delete(c.Auth.Config, clientIdKey)
	delete(c.Auth.Config, clientSecretKey)
	delete(c.Auth.Config, redirectUriKey)
	delete(c.Auth.Config, scopesKey)

	if pat != "" {
		c.Auth.Config[patKey] = pat
	}
	return pat != ""
}

func (c Config) ConfigureLoginAuth(clientId string, redirectUri string, scopes string) bool {
	delete(c.Auth.Config, clientSecretKey)
	delete(c.Auth.Config, patKey)

	if clientId != "" {
		c.Auth.Config[clientIdKey] = clientId
	}
	if redirectUri != "" {
		c.Auth.Config[redirectUriKey] = redirectUri
	}
	if scopes != "" {
		c.Auth.Config[scopesKey] = scopes
	}

	return clientId != "" || redirectUri != "" || scopes != ""
}

func (c Config) ConfigureCredentialsAuth(clientId string, clientSecret string) bool {
	delete(c.Auth.Config, redirectUriKey)
	delete(c.Auth.Config, scopesKey)
	delete(c.Auth.Config, patKey)

	if clientId != "" {
		c.Auth.Config[clientIdKey] = clientId
	}
	if clientSecret != "" {
		c.Auth.Config[clientSecretKey] = clientSecret
	}
	return clientId != "" || clientSecret != ""
}

func (c *Config) SetUri(uri string) error {
	parsedUri, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("Invalid value for 'uri': %w", err)
	}
	c.Uri = parsedUri
	return nil
}

func (c *Config) SetInsecure(insecure bool) {
	c.Insecure = insecure
}

func (c *Config) SetDebug(debug bool) {
	c.Debug = debug
}

func (c Config) SetHeader(key string, value string) {
	c.Header[key] = value
}

func (c Config) SetPath(key string, value string) {
	c.Path[key] = value
}

func (c Config) SetQuery(key string, value string) {
	c.Query[key] = value
}

func (c Config) SetAuthGrantType(grantType string) {
	c.Auth.Config["grantType"] = grantType
}

func (c Config) SetAuthScopes(scopes string) {
	c.Auth.Config["scopes"] = scopes
}

func (c Config) SetAuthProperty(key string, value string) {
	properties, ok := c.Auth.Config["properties"].(map[interface{}]interface{})
	if properties == nil || !ok {
		properties = map[interface{}]interface{}{}
	}
	properties[key] = value
	c.Auth.Config["properties"] = properties
}
