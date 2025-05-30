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

func (e PluginExecutor) authenticatorContext(ctx ExecutionContext, logger log.Logger) auth.AuthenticatorContext {
	authRequest := *auth.NewAuthenticatorRequest(ctx.BaseUri.String(), map[string]string{})
	return *auth.NewAuthenticatorContext(
		ctx.AuthConfig,
		ctx.IdentityUri,
		ctx.Settings.OperationId,
		ctx.Settings.Insecure,
		ctx.Debug,
		authRequest,
		logger)
}

func (e PluginExecutor) executeAuthenticators(ctx ExecutionContext, logger log.Logger) (*auth.AuthenticatorResult, error) {
	var token *auth.AuthToken = nil
	for _, authProvider := range e.authenticators {
		authContext := e.authenticatorContext(ctx, logger)
		result := authProvider.Auth(authContext)
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		if result.Token != nil {
			token = result.Token
		}
	}
	return auth.AuthenticatorSuccess(token), nil
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
		Token: auth.Token,
	}
}

func (e PluginExecutor) Call(ctx ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	auth, err := e.executeAuthenticators(ctx, logger)
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
		ctx.IdentityUri,
		ctx.Input,
		pluginParams,
		ctx.Debug,
		*plugin.NewExecutionSettings(ctx.Settings.OperationId, ctx.Settings.Header, ctx.Settings.Timeout, ctx.Settings.MaxAttempts, ctx.Settings.Insecure))
	return ctx.Plugin.Execute(*pluginContext, writer, logger)
}

func NewPluginExecutor(authenticators []auth.Authenticator) *PluginExecutor {
	return &PluginExecutor{authenticators}
}
