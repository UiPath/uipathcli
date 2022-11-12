package auth

import (
	"fmt"
	"strings"

	"github.com/UiPath/uipathcli/cache"
)

type TokenAuthenticator struct {
	Cache cache.Cache
}

func (a TokenAuthenticator) CanAuthenticate(ctx AuthenticatorContext) bool {
	 return strings.EqualFold("bearer", ctx.Type)
}

func (a TokenAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	config, err := a.getConfig(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid bearer authenticator configuration: %v", err))
	}

	ctx.Request.Header["Authorization"] = "Bearer " + config.Token
	return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
}

func (a TokenAuthenticator) enabled(ctx AuthenticatorContext) bool {
	return ctx.Config["clientId"] != nil
}

func (a TokenAuthenticator) getConfig(ctx AuthenticatorContext) (*TokenAuthenticatorConfig, error) {
	token, err := a.parseRequiredString(ctx.Config, "token")
	if err != nil {
		return nil, err
	}

	return NewTokenAuthenticatorConfig(token), nil
}

func (a TokenAuthenticator) parseRequiredString(config map[string]interface{}, name string) (string, error) {
	value := config[name]
	result, valid := value.(string)
	if !valid || result == "" {
		return "", fmt.Errorf("Invalid value for %s: '%v'", name, value)
	}
	return result, nil
}
