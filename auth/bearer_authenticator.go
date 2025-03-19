package auth

import (
	"fmt"
	"os"

	"github.com/UiPath/uipathcli/cache"
)

const ClientIdEnvVarName = "UIPATH_CLIENT_ID"
const ClientSecretEnvVarName = "UIPATH_CLIENT_SECRET" //nolint // This is not a secret but just the env variable name

// The BearerAuthenticator calls the identity token-endpoint to retrieve a JWT bearer token.
// It requires clientId and clientSecret.
type BearerAuthenticator struct {
	cache cache.Cache
}

func (a BearerAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	if !a.enabled(ctx) {
		return *AuthenticatorSuccess(nil)
	}
	config, err := a.getConfig(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid bearer authenticator configuration: %w", err))
	}
	identityClient := newIdentityClient(a.cache)
	tokenRequest := newTokenRequest(
		config.IdentityUri,
		config.GrantType,
		config.Scopes,
		config.ClientId,
		config.ClientSecret,
		config.Properties,
		ctx.OperationId,
		ctx.Insecure)
	tokenResponse, err := identityClient.GetToken(*tokenRequest)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error retrieving bearer token: %w", err))
	}
	return *AuthenticatorSuccess(NewBearerToken(tokenResponse.AccessToken))
}

func (a BearerAuthenticator) enabled(ctx AuthenticatorContext) bool {
	clientIdSet := os.Getenv(ClientIdEnvVarName) != "" || ctx.Config["clientId"] != nil
	clientSecretSet := os.Getenv(ClientSecretEnvVarName) != "" || ctx.Config["clientSecret"] != nil
	return clientIdSet && clientSecretSet
}

func (a BearerAuthenticator) getConfig(ctx AuthenticatorContext) (*bearerAuthenticatorConfig, error) {
	grantType, err := a.parseString(ctx.Config, "grantType")
	if err != nil {
		return nil, err
	}
	if grantType == "" {
		grantType = "client_credentials"
	}
	scopes, err := a.parseString(ctx.Config, "scopes")
	if err != nil {
		return nil, err
	}
	clientId, err := a.parseRequiredString(ctx.Config, "clientId", os.Getenv(ClientIdEnvVarName))
	if err != nil {
		return nil, err
	}
	clientSecret, err := a.parseRequiredString(ctx.Config, "clientSecret", os.Getenv(ClientSecretEnvVarName))
	if err != nil {
		return nil, err
	}
	properties, err := a.parseProperties(ctx.Config)
	if err != nil {
		return nil, err
	}
	return newBearerAuthenticatorConfig(grantType, scopes, clientId, clientSecret, properties, ctx.IdentityUri), nil
}

func (a BearerAuthenticator) parseProperties(config map[string]interface{}) (map[string]string, error) {
	result := map[string]string{}
	value := config["properties"]
	if value == nil {
		return result, nil
	}
	properties, valid := value.(map[interface{}]interface{})
	if !valid {
		return result, fmt.Errorf("Invalid key 'properties' in auth")
	}

	for k, v := range properties {
		key, valid := k.(string)
		if !valid {
			return result, fmt.Errorf("Invalid key '%v' in auth properties", k)
		}
		value, valid := v.(string)
		if !valid {
			return result, fmt.Errorf("Invalid value for '%v' in auth properties", k)
		}
		result[key] = value
	}
	return result, nil
}

func (a BearerAuthenticator) parseString(config map[string]interface{}, name string) (string, error) {
	value := config[name]
	result, valid := value.(string)
	if value != nil && !valid {
		return "", fmt.Errorf("Invalid value for %s: '%v'", name, value)
	}
	return result, nil
}

func (a BearerAuthenticator) parseRequiredString(config map[string]interface{}, name string, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	value := config[name]
	result, valid := value.(string)
	if !valid || result == "" {
		return "", fmt.Errorf("Invalid value for %s: '%v'", name, value)
	}
	return result, nil
}

func NewBearerAuthenticator(cache cache.Cache) *BearerAuthenticator {
	return &BearerAuthenticator{cache}
}
