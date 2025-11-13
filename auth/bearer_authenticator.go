package auth

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/utils/network"
)

const GetTokenTimeout = time.Duration(60) * time.Second
const GetTokenMaxAttempts = 3

const TokenExpiryGracePeriod = time.Duration(2) * time.Minute

const ClientIdEnvVarName = "UIPATH_CLIENT_ID"
const ClientSecretEnvVarName = "UIPATH_CLIENT_SECRET" //nolint:gosec // This is not a secret but just the env variable name
const GrantTypeEnvVarName = "UIPATH_AUTH_GRANT_TYPE"
const ScopesEnvVarName = "UIPATH_AUTH_SCOPES"

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

	tokenRequest := newTokenRequest(
		config.IdentityUri,
		config.GrantType,
		config.Scopes,
		config.ClientId,
		config.ClientSecret,
		config.Properties,
		a.networkSettings(ctx))

	tokenResponse := a.getAccessTokenFromCache(*tokenRequest)
	if tokenResponse != nil {
		ctx.Logger.Log(fmt.Sprintf("Using existing access token from local cache which expires at %s\n", tokenResponse.ExpiresAt.UTC().Format(time.RFC3339)))
		return *AuthenticatorSuccess(NewBearerToken(tokenResponse.AccessToken))
	}

	ctx.Logger.Log("No access token available. Calling identity server to retrieve new token...\n")

	identityClient := newIdentityClient(ctx.Logger)
	tokenResponse, err = identityClient.GetToken(*tokenRequest)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error retrieving bearer token: %w", err))
	}
	a.updateTokenResponseCache(*tokenRequest, *tokenResponse)
	return *AuthenticatorSuccess(NewBearerToken(tokenResponse.AccessToken))
}

func (a BearerAuthenticator) getAccessTokenFromCache(tokenRequest tokenRequest) *tokenResponse {
	cacheKey := a.cacheKey(tokenRequest)
	token, expiresAt := a.cache.Get(cacheKey)
	if token == "" {
		return nil
	}
	return newTokenResponse(token, expiresAt, nil)
}

func (a BearerAuthenticator) updateTokenResponseCache(tokenRequest tokenRequest, tokenResponse tokenResponse) {
	cacheKey := a.cacheKey(tokenRequest)
	a.cache.Set(cacheKey, tokenResponse.AccessToken, tokenResponse.ExpiresAt.Add(-TokenExpiryGracePeriod))
}

func (a BearerAuthenticator) cacheKey(tokenRequest tokenRequest) string {
	return fmt.Sprintf("beareraccesstoken|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
		tokenRequest.BaseUri.Scheme,
		tokenRequest.BaseUri.Hostname(),
		tokenRequest.GrantType,
		tokenRequest.Scopes,
		tokenRequest.ClientId,
		tokenRequest.ClientSecret,
		tokenRequest.Code,
		tokenRequest.CodeVerifier,
		tokenRequest.RedirectUri,
		a.cacheKeyProperties(tokenRequest.Properties))
}

func (a BearerAuthenticator) cacheKeyProperties(properties map[string]string) string {
	values := []string{}
	for key, value := range properties {
		values = append(values, key+"="+value)
	}
	return strings.Join(values, ",")
}

func (a BearerAuthenticator) enabled(ctx AuthenticatorContext) bool {
	clientIdSet := os.Getenv(ClientIdEnvVarName) != "" || ctx.Config["clientId"] != nil
	clientSecretSet := os.Getenv(ClientSecretEnvVarName) != "" || ctx.Config["clientSecret"] != nil
	isOAuthFlow := os.Getenv(RedirectUriVarName) != "" || ctx.Config["redirectUri"] != nil
	return clientIdSet && clientSecretSet && !isOAuthFlow
}

func (a BearerAuthenticator) getConfig(ctx AuthenticatorContext) (*bearerAuthenticatorConfig, error) {
	grantType, err := a.parseString(ctx.Config, "grantType", GrantTypeEnvVarName)
	if err != nil {
		return nil, err
	}
	if grantType == "" {
		grantType = "client_credentials"
	}
	scopes, err := a.parseString(ctx.Config, "scopes", ScopesEnvVarName)
	if err != nil {
		return nil, err
	}
	clientId, err := a.parseRequiredString(ctx.Config, "clientId", ClientIdEnvVarName)
	if err != nil {
		return nil, err
	}
	clientSecret, err := a.parseRequiredString(ctx.Config, "clientSecret", ClientSecretEnvVarName)
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
	properties, valid := value.(map[string]interface{})
	if !valid {
		return result, errors.New("Invalid key 'properties' in auth")
	}

	for k, v := range properties {
		value, valid := v.(string)
		if !valid {
			return result, fmt.Errorf("Invalid value for '%s' in auth properties", k)
		}
		result[k] = value
	}
	return result, nil
}

func (a BearerAuthenticator) parseString(config map[string]interface{}, name string, envVarName string) (string, error) {
	envVarValue := os.Getenv(envVarName)
	if envVarValue != "" {
		return envVarValue, nil
	}
	value := config[name]
	result, valid := value.(string)
	if value != nil && !valid {
		return "", fmt.Errorf("Invalid value for %s: '%v'", name, value)
	}
	return result, nil
}

func (a BearerAuthenticator) parseRequiredString(config map[string]interface{}, name string, envVarName string) (string, error) {
	envVarValue := os.Getenv(envVarName)
	if envVarValue != "" {
		return envVarValue, nil
	}
	value := config[name]
	result, valid := value.(string)
	if !valid || result == "" {
		return "", fmt.Errorf("Invalid value for %s: '%v'", name, value)
	}
	return result, nil
}

func (a BearerAuthenticator) networkSettings(ctx AuthenticatorContext) network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		ctx.Debug,
		ctx.OperationId,
		map[string]string{},
		GetTokenTimeout,
		GetTokenMaxAttempts,
		ctx.Insecure,
	)
}

func NewBearerAuthenticator(cache cache.Cache) *BearerAuthenticator {
	return &BearerAuthenticator{cache}
}
