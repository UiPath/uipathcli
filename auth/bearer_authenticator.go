package auth

import (
	"fmt"
	"net/url"
	"os"

	"github.com/UiPath/uipathcli/cache"
)

const ClientIdEnvVarName = "UIPATH_CLIENT_ID"
const ClientSecretEnvVarName = "UIPATH_CLIENT_SECRET"
const IdentityUriEnvVarName = "UIPATH_IDENTITY_URI"

// The BearerAuthenticator calls the identity token-endpoint to retrieve a JWT bearer token.
// It requires clientId and clientSecret.
type BearerAuthenticator struct {
	Cache cache.Cache
}

func (a BearerAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	if !a.enabled(ctx) {
		return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
	}
	config, err := a.getConfig(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid bearer authenticator configuration: %v", err))
	}
	identityBaseUri := config.IdentityUri
	if identityBaseUri == nil {
		requestUrl, err := url.Parse(ctx.Request.URL)
		if err != nil {
			return *AuthenticatorError(fmt.Errorf("Invalid request url '%s': %v", ctx.Request.URL, err))
		}
		identityBaseUri, err = url.Parse(fmt.Sprintf("%s://%s/identity_", requestUrl.Scheme, requestUrl.Host))
		if err != nil {
			return *AuthenticatorError(fmt.Errorf("Invalid identity url '%s': %v", ctx.Request.URL, err))
		}
	}

	identityClient := identityClient(a)
	tokenRequest := newTokenRequest(
		*identityBaseUri,
		config.GrantType,
		config.Scopes,
		config.ClientId,
		config.ClientSecret,
		config.Properties,
		ctx.Insecure)
	tokenResponse, err := identityClient.GetToken(*tokenRequest)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error retrieving bearer token: %v", err))
	}
	ctx.Request.Header["Authorization"] = "Bearer " + tokenResponse.AccessToken
	return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
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
	var uri *url.URL
	uriString, err := a.parseRequiredString(ctx.Config, "uri", os.Getenv(IdentityUriEnvVarName))
	if err == nil {
		uri, err = url.Parse(uriString)
		if err != nil {
			return nil, fmt.Errorf("Error parsing identity uri: %v", err)
		}
	}
	return newBearerAuthenticatorConfig(grantType, scopes, clientId, clientSecret, properties, uri), nil
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
