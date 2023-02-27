package auth

import (
	"fmt"
	"os"
)

const PatEnvVarName = "UIPATH_PAT"

type PatAuthenticator struct{}

func (a PatAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	if !a.enabled(ctx) {
		return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
	}
	pat, err := a.getPat(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid PAT authenticator configuration: %v", err))
	}
	ctx.Request.Header["Authorization"] = "Bearer " + pat
	return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
}

func (a PatAuthenticator) enabled(ctx AuthenticatorContext) bool {
	return os.Getenv(PatEnvVarName) != "" || ctx.Config["pat"] != nil
}

func (a PatAuthenticator) getPat(ctx AuthenticatorContext) (string, error) {
	return a.parseRequiredString(ctx.Config, "pat", os.Getenv(PatEnvVarName))
}

func (a PatAuthenticator) parseRequiredString(config map[string]interface{}, name string, override string) (string, error) {
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
