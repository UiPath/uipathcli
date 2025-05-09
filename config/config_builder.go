package config

type ConfigBuilder struct {
	Config Config

	organization *string
	tenant       *string

	authType string

	clientId     *string
	clientSecret *string

	redirectUri *string
	scopes      *string

	pat *string
}

func (b *ConfigBuilder) WithOrganization(organization *string) *ConfigBuilder {
	b.organization = organization
	return b
}

func (b *ConfigBuilder) WithTenant(tenant *string) *ConfigBuilder {
	b.tenant = tenant
	return b
}

func (b *ConfigBuilder) WithCredentials(clientId *string, clientSecret *string) *ConfigBuilder {
	b.authType = AuthTypeCredentials
	b.clientId = clientId
	b.clientSecret = clientSecret
	return b
}

func (b *ConfigBuilder) WithLogin(clientId *string, clientSecret *string, redirectUri *string, scopes *string) *ConfigBuilder {
	b.authType = AuthTypeLogin
	b.clientId = clientId
	b.clientSecret = clientSecret
	b.redirectUri = redirectUri
	b.scopes = scopes
	return b
}

func (b *ConfigBuilder) WithPat(pat *string) *ConfigBuilder {
	b.authType = AuthTypePat
	b.pat = pat
	return b
}

func (b *ConfigBuilder) Build() (Config, bool) {
	cfg := b.Config
	authType := cfg.AuthType()

	if b.organization != nil {
		cfg.SetOrganization(*b.organization)
	}
	if b.tenant != nil {
		cfg.SetTenant(*b.tenant)
	}

	switch b.authType {
	case AuthTypeCredentials:
		cfg.SetCredentialsAuth(b.clientId, b.clientSecret)
	case AuthTypeLogin:
		cfg.SetLoginAuth(b.clientId, b.clientSecret, b.redirectUri, b.scopes)
	case AuthTypePat:
		cfg.SetPatAuth(b.pat)
	}

	changed := b.organization != nil ||
		b.tenant != nil ||
		b.clientId != nil ||
		b.clientSecret != nil ||
		b.redirectUri != nil ||
		b.scopes != nil ||
		b.pat != nil ||
		(b.authType != "" && b.authType != authType)
	return cfg, changed
}

func NewConfigBuilder(cfg Config) *ConfigBuilder {
	return &ConfigBuilder{
		Config: cfg,
	}
}
