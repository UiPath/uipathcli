package dataservice

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/commandline"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/api"
	"github.com/UiPath/uipathcli/utils/converter"
)

// The DataServiceCommand implements dynamic CLI commands for the
// dataservice entites. It discovers the existing entites from the
// service and creates for each entity a set of CRUD commands.
type DataServiceCommand struct {
	args           []string
	stdErr         io.Writer
	configProvider config.ConfigProvider
	authenticators []auth.Authenticator
}

func (c DataServiceCommand) Commands() []plugin.Command {
	commands := []plugin.Command{}

	operationId := c.operationId()
	definition, err := c.getEntityDefinition(operationId)
	if err != nil {
		_, _ = fmt.Fprintf(c.stdErr, "Error retrieving entity specification from dataservice: %v\n", err)
		return commands
	}

	for _, operation := range definition.Operations {
		entityName := operation.Category.Name
		name := c.formatOperationName(operation.Name, entityName)
		command := plugin.NewCommand("dataservice").
			WithCategory(entityName, operation.Category.Summary, operation.Category.Description).
			WithOperation(name, operation.Summary, operation.Description)
		for _, parameter := range operation.Parameters {
			command.WithParameter(parameter.Name, parameter.Type, parameter.Description, parameter.Required)
		}
		commands = append(commands, *command)
	}
	return commands
}

func (c DataServiceCommand) formatOperationName(name string, entityName string) string {
	result := strings.ReplaceAll(name, " ", "-")
	result = strings.ReplaceAll(result, "-to-"+entityName, "")
	result = strings.ReplaceAll(result, "-from-"+entityName, "")
	result = strings.ReplaceAll(result, "-"+entityName, "")
	return result
}

func (c DataServiceCommand) Command() plugin.Command {
	return *plugin.NewCommand("dataservice")
}

func (c DataServiceCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if len(c.args) < 4 {
		return nil
	}
	entityName := c.args[2]
	operationName := c.formatOperationName(c.args[3], entityName)

	definition, err := c.getEntityDefinition(ctx.Settings.OperationId)
	if err != nil {
		return fmt.Errorf("Error retrieving entity specification from dataservice: %w", err)
	}

	parameterValues := map[string]interface{}{}
	for _, parameter := range ctx.Parameters {
		parameterValues[parameter.Name] = parameter.Value
	}

	operation := c.findOperation(definition, entityName, operationName)
	if operation == nil {
		return fmt.Errorf("Could not find operation '%s' for entity '%s' in dataservice", operationName, entityName)
	}

	uriBuilder := converter.NewUriBuilder(ctx.BaseUri, operation.Route)
	uriBuilder.AddQueryString("organization", ctx.Organization)
	uriBuilder.AddQueryString("tenant", ctx.Tenant)
	data := map[string]interface{}{}
	for _, parameter := range operation.Parameters {
		value := parameterValues[parameter.Name]
		if value == nil {
			continue
		}
		if parameter.In == parser.ParameterInPath {
			uriBuilder.FormatPath(parameter.Name, value)
		}
		if parameter.In == parser.ParameterInQuery {
			uriBuilder.AddQueryString(parameter.FieldName, value)
		}
		if parameter.In == parser.ParameterInBody {
			data[parameter.FieldName] = value
		}
	}

	uri := uriBuilder.Build()

	client := api.NewDataServiceClient(ctx.BaseUri.String(), ctx.Auth.Token, ctx.Debug, ctx.Settings, logger)
	response, err := client.CallEntity(operation.Method, uri, map[string]string{}, data)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Error reading response from Data Service: %w", err)
	}
	if response.StatusCode > 299 {
		return fmt.Errorf("Data Service returned status code '%v' and body '%v'", response.StatusCode, string(responseBody))
	}
	return writer.WriteResponse(*output.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(responseBody)))
}

func (c DataServiceCommand) getStringValue(argsParser *commandline.ArgsParser, flagName string, defaultValue string) string {
	value := argsParser.GetValue(flagName)
	stringValue, ok := value.(string)
	if ok {
		return stringValue
	}
	return defaultValue
}

func (c DataServiceCommand) getUriValue(argsParser *commandline.ArgsParser, flagName string, defaultValue *url.URL) (*url.URL, error) {
	value := argsParser.GetValue(flagName)
	stringValue, ok := value.(string)
	if ok {
		return url.Parse(stringValue)
	}
	return defaultValue, nil
}

func (c DataServiceCommand) getEntityDefinition(operationId string) (*parser.Definition, error) {
	flags := commandline.NewFlagBuilder().
		AddDefaultFlags(false).
		AddHelpFlag().
		AddVersionFlag().
		Build()
	argsParser, err := commandline.NewArgsParser(c.args, flags)
	if err != nil {
		return nil, err
	}

	profile := c.getStringValue(argsParser, commandline.FlagNameProfile, config.DefaultProfile)
	config := c.configProvider.Config(profile)

	organization := c.getStringValue(argsParser, commandline.FlagNameOrganization, config.Organization)
	tenant := c.getStringValue(argsParser, commandline.FlagNameTenant, config.Tenant)

	defaultBaseUri, _ := url.Parse("https://cloud.uipath.com/{organization}/{tenant}/dataservice_")
	if config.Uri != nil {
		defaultBaseUri = config.Uri
	}
	baseUri, err := c.getUriValue(argsParser, commandline.FlagNameUri, defaultBaseUri)
	if err != nil {
		return nil, err
	}

	defaultIdentityUri, _ := url.Parse(fmt.Sprintf("%s://%s/identity_", baseUri.Scheme, baseUri.Host))
	configIdentityUrl, ok := config.Auth.Config["uri"].(string)
	if ok {
		defaultIdentityUri, err = url.Parse(configIdentityUrl)
		if err != nil {
			return nil, err
		}
	}
	identityUri, err := c.getUriValue(argsParser, commandline.FlagNameIdentityUri, defaultIdentityUri)
	if err != nil {
		return nil, err
	}

	specification, err := c.getEntitySpecification(config.Auth, *baseUri, organization, tenant, *identityUri, operationId, false)
	if err != nil {
		return nil, err
	}
	p := parser.NewOpenApiParser()
	return p.Parse("dataservice", specification)
}

func (c DataServiceCommand) getEntitySpecification(authConfig config.AuthConfig, baseUri url.URL, organization string, tenant string, identityUri url.URL, operationId string, insecure bool) ([]byte, error) {
	// TODO: add caching
	return c.downloadEntitySpecification(authConfig, baseUri, organization, tenant, identityUri, operationId, insecure)
}

func (c DataServiceCommand) downloadEntitySpecification(authConfig config.AuthConfig, baseUri url.URL, organization string, tenant string, identityUri url.URL, operationId string, insecure bool) ([]byte, error) {
	auth, err := c.executeAuthenticators(authConfig, baseUri.String(), identityUri, operationId, insecure)
	if err != nil {
		return nil, err
	}

	uri := c.formatUri(baseUri, organization, tenant)
	clientSettings := plugin.NewExecutionSettings(operationId, map[string]string{}, time.Duration(60)*time.Second, 3, insecure)
	client := api.NewDataServiceClient(uri, auth.Token, false, *clientSettings, nil)
	return client.GetEntitySpecification()
}

func (c DataServiceCommand) executeAuthenticators(authConfig config.AuthConfig, baseUri string, identityUri url.URL, operationId string, insecure bool) (*auth.AuthenticatorResult, error) {
	authRequest := *auth.NewAuthenticatorRequest(baseUri, map[string]string{})
	ctx := *auth.NewAuthenticatorContext(authConfig.Type, authConfig.Config, identityUri, operationId, insecure, authRequest)
	var token *auth.AuthToken = nil
	for _, authProvider := range c.authenticators {
		result := authProvider.Auth(ctx)
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		if result.Token != nil {
			token = result.Token
		}
	}
	return auth.AuthenticatorSuccess(token), nil
}

func (c DataServiceCommand) findOperation(definition *parser.Definition, entityName string, operationName string) *parser.Operation {
	for _, operation := range definition.Operations {
		if entityName == operation.Category.Name && operationName == c.formatOperationName(operation.Name, operation.Category.Name) {
			return &operation
		}
	}
	return nil
}

func (c DataServiceCommand) formatUri(baseUri url.URL, org string, tenant string) string {
	path := baseUri.Path
	if baseUri.Path == "" {
		path = "/{organization}/{tenant}/dataservice_"
	}
	path = strings.ReplaceAll(path, "{organization}", org)
	path = strings.ReplaceAll(path, "{tenant}", tenant)
	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("%s://%s%s", baseUri.Scheme, baseUri.Host, path)
}

func (c DataServiceCommand) operationId() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("%x%x%x%x%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}

func NewDataServiceCommand(args []string, stdErr io.Writer, configProvider config.ConfigProvider, authenticators []auth.Authenticator) *DataServiceCommand {
	return &DataServiceCommand{args, stdErr, configProvider, authenticators}
}
