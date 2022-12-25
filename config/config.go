package config

import (
	"fmt"
	"net/url"
)

type Config struct {
	Uri      *url.URL
	Path     map[string]string
	Query    map[string]string
	Header   map[string]string
	Auth     AuthConfig
	Insecure bool
	Debug    bool
}

type AuthConfig struct {
	Type   string
	Config map[string]interface{}
}

const organizationKey = "organization"
const tenantKey = "tenant"

func (a Config) Organization() string {
	return a.Path[organizationKey]
}

func (a Config) Tenant() string {
	return a.Path[tenantKey]
}

func (a Config) ConfigureOrgTenant(organization string, tenant string) bool {
	if organization != "" {
		a.Path[organizationKey] = organization
	}
	if tenant != "" {
		a.Path[tenantKey] = tenant
	}

	return organization != "" || tenant != ""
}

const clientIdKey = "clientId"
const clientSecretKey = "clientSecret"
const redirectUriKey = "redirectUri"
const scopesKey = "scopes"

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

func (c Config) ConfigureLoginAuth(clientId string, redirectUri string, scopes string) bool {
	delete(c.Auth.Config, clientSecretKey)

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

	if clientId != "" {
		c.Auth.Config[clientIdKey] = clientId
	}
	if clientSecret != "" {
		c.Auth.Config[clientSecretKey] = clientSecret
	}
	return clientId != "" || clientSecret != ""
}
