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

const clientIdKey = "clientId"
const clientSecretKey = "clientSecret"
const redirectUriKey = "redirectUri"
const scopesKey = "scopes"
const patKey = "pat"

func (c Config) Organization() string {
	return c.Path[organizationKey]
}

func (c Config) Tenant() string {
	return c.Path[tenantKey]
}

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

func (c Config) ConfigureOrgTenant(organization string, tenant string) bool {
	if organization != "" {
		c.Path[organizationKey] = organization
	}
	if tenant != "" {
		c.Path[tenantKey] = tenant
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
