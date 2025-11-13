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
	Auth           map[string]interface{}
	Insecure       bool
	Debug          bool
	Output         string
	ServiceVersion string
}

const clientIdKey = "clientId"
const clientSecretKey = "clientSecret"
const redirectUriKey = "redirectUri"
const scopesKey = "scopes"
const patKey = "pat"
const grantTypeKey = "grantType"
const authUriKey = "uri"
const propertiesKey = "properties"

func (c *Config) ClientId() string {
	clientId := c.Auth[clientIdKey]
	if clientId == nil {
		return ""
	}
	return fmt.Sprint(clientId)
}

func (c *Config) ClientSecret() string {
	clientSecret := c.Auth[clientSecretKey]
	if clientSecret == nil {
		return ""
	}
	return fmt.Sprint(clientSecret)
}

func (c *Config) RedirectUri() string {
	redirectUri := c.Auth[redirectUriKey]
	if redirectUri == nil {
		return ""
	}
	return fmt.Sprint(redirectUri)
}

func (c *Config) Scopes() string {
	scopes := c.Auth[scopesKey]
	if scopes == nil {
		return ""
	}
	return fmt.Sprint(scopes)
}

func (c *Config) Pat() string {
	pat := c.Auth[patKey]
	if pat == nil {
		return ""
	}
	return fmt.Sprint(pat)
}

func (c *Config) AuthUri() string {
	identityUri := c.Auth[authUriKey]
	if identityUri == nil {
		return ""
	}
	return fmt.Sprint(identityUri)
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
	delete(c.Auth, redirectUriKey)
	delete(c.Auth, scopesKey)
	delete(c.Auth, patKey)

	if clientId != nil {
		c.Auth[clientIdKey] = *clientId
	}
	if clientSecret != nil {
		c.Auth[clientSecretKey] = *clientSecret
	}
}

func (c *Config) SetLoginAuth(clientId *string, clientSecret *string, redirectUri *string, scopes *string) {
	delete(c.Auth, patKey)

	if clientId != nil {
		c.Auth[clientIdKey] = *clientId
	}
	if clientSecret != nil {
		c.Auth[clientSecretKey] = *clientSecret
	}
	if redirectUri != nil {
		c.Auth[redirectUriKey] = *redirectUri
	}
	if scopes != nil {
		c.Auth[scopesKey] = *scopes
	}
}

func (c *Config) SetPatAuth(pat *string) {
	delete(c.Auth, clientIdKey)
	delete(c.Auth, clientSecretKey)
	delete(c.Auth, redirectUriKey)
	delete(c.Auth, scopesKey)

	if pat != nil {
		c.Auth[patKey] = *pat
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
	c.Auth[grantTypeKey] = grantType
}

func (c *Config) SetAuthScopes(scopes string) {
	c.Auth[scopesKey] = scopes
}

func (c *Config) SetAuthUri(uri string) error {
	parsedUri, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("Invalid value for 'auth.uri': %w", err)
	}
	c.Auth[authUriKey] = parsedUri.String()
	return nil
}

func (c *Config) SetAuthProperty(key string, value string) {
	properties, ok := c.Auth[propertiesKey].(map[string]interface{})
	if properties == nil || !ok {
		properties = map[string]interface{}{}
	}
	properties[key] = value
	c.Auth[propertiesKey] = properties
}

func (c *Config) SetServiceVersion(serviceVersion string) {
	c.ServiceVersion = serviceVersion
}
