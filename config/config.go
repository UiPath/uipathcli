// Package config parses config and plugin configuration files. It provides APIs to read
// and write config files and manage multiple profiles.
package config

import (
	"fmt"
	"net/url"
)

const AuthTypeCredentials = "credentials"
const AuthTypeLogin = "login"
const AuthTypePat = "pat"

// The Config structure holds the config data from the selected profile.
type Config struct {
	Uri            *url.URL
	Organization   string
	Tenant         string
	Parameter      map[string]string
	Header         map[string]string
	Auth           AuthConfig
	Insecure       bool
	Debug          bool
	Output         string
	ServiceVersion string
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

func (c *Config) ClientId() string {
	clientId := c.Auth.Config[clientIdKey]
	if clientId == nil {
		return ""
	}
	return fmt.Sprint(clientId)
}

func (c *Config) ClientSecret() string {
	clientSecret := c.Auth.Config[clientSecretKey]
	if clientSecret == nil {
		return ""
	}
	return fmt.Sprint(clientSecret)
}

func (c *Config) RedirectUri() string {
	redirectUri := c.Auth.Config[redirectUriKey]
	if redirectUri == nil {
		return ""
	}
	return fmt.Sprint(redirectUri)
}

func (c *Config) Scopes() string {
	scopes := c.Auth.Config[scopesKey]
	if scopes == nil {
		return ""
	}
	return fmt.Sprint(scopes)
}

func (c *Config) Pat() string {
	pat := c.Auth.Config[patKey]
	if pat == nil {
		return ""
	}
	return fmt.Sprint(pat)
}

func (c *Config) AuthType() string {
	if c.Pat() != "" {
		return AuthTypePat
	}
	if c.RedirectUri() != "" {
		return AuthTypeLogin
	}
	if c.ClientId() != "" && c.ClientSecret() != "" {
		return AuthTypeCredentials
	}
	return ""
}

func (c *Config) SetOrganization(organization string) {
	c.Organization = organization
}

func (c *Config) SetTenant(tenant string) {
	c.Tenant = tenant
}

func (c *Config) SetCredentialsAuth(clientId *string, clientSecret *string) {
	delete(c.Auth.Config, redirectUriKey)
	delete(c.Auth.Config, scopesKey)
	delete(c.Auth.Config, patKey)

	if clientId != nil {
		c.Auth.Config[clientIdKey] = clientId
	}
	if clientSecret != nil {
		c.Auth.Config[clientSecretKey] = clientSecret
	}
}

func (c *Config) SetLoginAuth(clientId *string, clientSecret *string, redirectUri *string, scopes *string) {
	delete(c.Auth.Config, patKey)

	if clientId != nil {
		c.Auth.Config[clientIdKey] = clientId
	}
	if clientSecret != nil {
		c.Auth.Config[clientSecretKey] = clientSecret
	}
	if redirectUri != nil {
		c.Auth.Config[redirectUriKey] = redirectUri
	}
	if scopes != nil {
		c.Auth.Config[scopesKey] = scopes
	}
}

func (c *Config) SetPatAuth(pat *string) {
	delete(c.Auth.Config, clientIdKey)
	delete(c.Auth.Config, clientSecretKey)
	delete(c.Auth.Config, redirectUriKey)
	delete(c.Auth.Config, scopesKey)

	if pat != nil {
		c.Auth.Config[patKey] = pat
	}
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

func (c *Config) SetHeader(key string, value string) {
	c.Header[key] = value
}

func (c *Config) SetParameter(key string, value string) {
	c.Parameter[key] = value
}

func (c *Config) SetAuthGrantType(grantType string) {
	c.Auth.Config["grantType"] = grantType
}

func (c *Config) SetAuthScopes(scopes string) {
	c.Auth.Config["scopes"] = scopes
}

func (c *Config) SetAuthProperty(key string, value string) {
	properties, ok := c.Auth.Config["properties"].(map[interface{}]interface{})
	if properties == nil || !ok {
		properties = map[interface{}]interface{}{}
	}
	properties[key] = value
	c.Auth.Config["properties"] = properties
}

func (c *Config) SetServiceVersion(serviceVersion string) {
	c.ServiceVersion = serviceVersion
}
