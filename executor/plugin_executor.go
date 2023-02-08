package executor

import (
	"errors"
	"net/url"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
)

type PluginExecutor struct {
	Authenticators []auth.Authenticator
}

func (e PluginExecutor) executeAuthenticators(baseUri url.URL, authConfig config.AuthConfig, debug bool, insecure bool) (*auth.AuthenticatorResult, error) {
	authRequest := *auth.NewAuthenticatorRequest(baseUri.String(), map[string]string{})
	ctx := *auth.NewAuthenticatorContext(authConfig.Type, authConfig.Config, debug, insecure, authRequest)
	for _, authProvider := range e.Authenticators {
		result := authProvider.Auth(ctx)
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		ctx.Config = result.Config
		for k, v := range result.RequestHeader {
			ctx.Request.Header[k] = v
		}
	}
	return auth.AuthenticatorSuccess(ctx.Request.Header, ctx.Config), nil
}

func (e PluginExecutor) convertToPluginParameters(parameters []ExecutionParameter) []plugin.ExecutionParameter {
	result := []plugin.ExecutionParameter{}
	for _, parameter := range parameters {
		name := parameter.Name
		value := parameter.Value
		if fileReference, ok := parameter.Value.(FileReference); ok {
			value = *plugin.NewFileParameter(fileReference.path, fileReference.filename, fileReference.data)
		}
		result = append(result, *plugin.NewExecutionParameter(name, value))
	}
	return result
}

func (e PluginExecutor) pluginParameters(context ExecutionContext) []plugin.ExecutionParameter {
	params := context.PathParameters
	params = append(params, context.QueryParameters...)
	params = append(params, context.HeaderParameters...)
	params = append(params, context.BodyParameters...)
	params = append(params, context.FormParameters...)
	return e.convertToPluginParameters(params)
}

func (e PluginExecutor) pluginAuth(auth *auth.AuthenticatorResult) plugin.AuthResult {
	return plugin.AuthResult{
		Header: auth.RequestHeader,
	}
}

func (e PluginExecutor) Call(context ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	auth, err := e.executeAuthenticators(context.BaseUri, context.AuthConfig, context.Debug, context.Insecure)
	if err != nil {
		return err
	}

	var pluginInput *plugin.FileParameter
	if context.Input != nil {
		pluginInput = plugin.NewFileParameter(context.Input.path, context.Input.filename, context.Input.data)
	}
	pluginAuth := e.pluginAuth(auth)
	pluginParams := e.pluginParameters(context)
	pluginContext := plugin.NewExecutionContext(
		context.BaseUri,
		pluginAuth,
		pluginInput,
		pluginParams,
		context.Insecure,
		context.Debug)
	return context.Plugin.Execute(*pluginContext, writer, logger)
}
