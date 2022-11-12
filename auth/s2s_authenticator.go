package auth

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/UiPath/uipathcli/cache"
)

type S2SAuthenticator struct {
	Cache cache.Cache
}

func (a S2SAuthenticator) CanAuthenticate(ctx AuthenticatorContext) bool {
	 return strings.EqualFold("client_credentials", ctx.Type)
}

func (a S2SAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	if !a.enabled(ctx) {
		return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
	}
	config, err := a.getConfig(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid s2s authenticator configuration: %v", err))
	}

	url, err := url.Parse(ctx.Request.URL)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid request url '%s': %v", ctx.Request.URL, err))
	}

	identityClient := identityClient(a)
	tokenRequest := newTokenRequest(
		fmt.Sprintf("%s://%s/identity_", url.Scheme, url.Host),
		config.ClientId,
		config.ClientSecret,
		ctx.Insecure)
	token, err := identityClient.GetToken(*tokenRequest)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error retrieving s2s token: %v", err))
	}
	ctx.Request.Header["Authorization"] = "Bearer " + token
	return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
}

func (a S2SAuthenticator) enabled(ctx AuthenticatorContext) bool {
	return ctx.Config["clientId"] != nil && ctx.Config["clientSecret"] != nil
}

func (a S2SAuthenticator) getConfig(ctx AuthenticatorContext) (*S2SAuthenticatorConfig, error) {
	clientId, err := a.parseRequiredString(ctx.Config, "clientId")
	if err != nil {
		return nil, err
	}
	clientSecret, err := a.parseRequiredString(ctx.Config, "clientSecret")
	if err != nil {
		return nil, err
	}
	return NewS2SAuthenticatorConfig(clientId, clientSecret), nil
}

func (a S2SAuthenticator) parseRequiredString(config map[string]interface{}, name string) (string, error) {
	value := config[name]
	result, valid := value.(string)
	if !valid || result == "" {
		return "", fmt.Errorf("Invalid value for %s: '%v'", name, value)
	}
	return result, nil
}
