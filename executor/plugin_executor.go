package executor

import (
	"errors"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
)

// The PluginExecutor implements the Executor interface and invokes the
// registered plugin for the executed command.
// The plugin takes care of sending the HTTP request or performing other
// operations.
type PluginExecutor struct {
	authenticators []auth.Authenticator
}

func (e PluginExecutor) authenticatorContext(ctx ExecutionContext) auth.AuthenticatorContext {
	authRequest := *auth.NewAuthenticatorRequest(ctx.BaseUri.String(), map[string]string{})
	return *auth.NewAuthenticatorContext(
		ctx.AuthConfig.Type,
		ctx.AuthConfig.Config,
		ctx.IdentityUri,
		ctx.Settings.OperationId,
		ctx.Settings.Insecure,
		authRequest)
}

func (e PluginExecutor) executeAuthenticators(ctx ExecutionContext) (*auth.AuthenticatorResult, error) {
	authContext := e.authenticatorContext(ctx)
	for _, authProvider := range e.authenticators {
		result := authProvider.Auth(authContext)
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		authContext.Config = result.Config
		for k, v := range result.RequestHeader {
			authContext.Request.Header[k] = v
		}
	}
	return auth.AuthenticatorSuccess(authContext.Request.Header, authContext.Config), nil
}

func (e PluginExecutor) convertToPluginParameters(parameters []ExecutionParameter) []plugin.ExecutionParameter {
	result := []plugin.ExecutionParameter{}
	for _, parameter := range parameters {
		param := plugin.NewExecutionParameter(parameter.Name, parameter.Value)
		result = append(result, *param)
	}
	return result
}

func (e PluginExecutor) pluginAuth(auth *auth.AuthenticatorResult) plugin.AuthResult {
	return plugin.AuthResult{
		Header: auth.RequestHeader,
	}
}

func (e PluginExecutor) Call(ctx ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	auth, err := e.executeAuthenticators(ctx)
	if err != nil {
		return err
	}

	pluginAuth := e.pluginAuth(auth)
	pluginParams := e.convertToPluginParameters(ctx.Parameters)
	pluginContext := plugin.NewExecutionContext(
		ctx.Organization,
		ctx.Tenant,
		ctx.BaseUri,
		pluginAuth,
		ctx.Input,
		pluginParams,
		ctx.Debug,
		*plugin.NewExecutionSettings(ctx.Settings.OperationId, ctx.Settings.Timeout, ctx.Settings.MaxAttempts, ctx.Settings.Insecure))
	return ctx.Plugin.Execute(*pluginContext, writer, logger)
}

func NewPluginExecutor(authenticators []auth.Authenticator) *PluginExecutor {
	return &PluginExecutor{authenticators}
}
