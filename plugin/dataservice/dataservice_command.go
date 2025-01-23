package dataservice

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/converter"
	"github.com/UiPath/uipathcli/utils/resiliency"
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

	definition, err := c.getEntityDefinition()
	if err != nil {
		fmt.Fprintln(c.stdErr, fmt.Sprintf("Error retrieving entity specification from dataservice: %v", err))
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

func (c DataServiceCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if len(c.args) < 4 {
		return nil
	}
	entityName := c.args[2]
	operationName := c.formatOperationName(c.args[3], entityName)

	definition, err := c.getEntityDefinition()
	if err != nil {
		return fmt.Errorf("Error retrieving entity specification from dataservice: %v", err)
	}

	parameterValues := map[string]interface{}{}
	for _, parameter := range context.Parameters {
		parameterValues[parameter.Name] = parameter.Value
	}

	operation := c.findOperation(definition, entityName, operationName)
	if operation == nil {
		return fmt.Errorf("Could not find operation '%s' for entity '%s' in dataservice", operationName, entityName)
	}

	uriBuilder := converter.NewUriBuilder(context.BaseUri, operation.Route)
	uriBuilder.AddQueryString("organization", context.Organization)
	uriBuilder.AddQueryString("tenant", context.Tenant)
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
	request, err := c.createRequest(operation.Method, uri, context.Auth.Header, data)
	if err != nil {
		return fmt.Errorf("Error creating request: %w", err)
	}
	if context.Debug {
		c.logRequest(logger, request)
	}
	response, err := c.sendRequest(request, context.Insecure)
	if err != nil {
		return fmt.Errorf("Error sending request: %w", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Error reading response: %w", err)
	}
	c.logResponse(logger, response, responseBody)
	return writer.WriteResponse(*output.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(responseBody)))
}

func (c DataServiceCommand) createRequest(method string, uri string, header map[string]string, body map[string]interface{}) (*http.Request, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("Error creating body: %w", err)
	}
	request, err := http.NewRequest(method, uri, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	request.Header.Add("content-type", "application/json")
	for key, value := range header {
		request.Header.Add(key, value)
	}
	return request, nil
}

func (c DataServiceCommand) getEntityDefinition() (*parser.Definition, error) {
	baseUri, err := url.Parse("https://cloud.uipath.com/{organization}/{tenant}/dataservice_/api")
	if err != nil {
		return nil, err
	}
	identityUri, err := url.Parse("https://cloud.uipath.com/identity_")
	if err != nil {
		return nil, err
	}
	specification, err := c.getEntitySpecification(*baseUri, "uipatcleitzc", "DefaultTenant", *identityUri, false, false)
	if err != nil {
		return nil, err
	}
	p := parser.NewOpenApiParser()
	return p.Parse("dataservice", specification)
}

func (c DataServiceCommand) getEntitySpecification(baseUri url.URL, organization string, tenant string, identityUri url.URL, debug bool, insecure bool) ([]byte, error) {
	var response []byte
	err := resiliency.Retry(func() error {
		var err error
		response, err = c.downloadEntitySpecification(baseUri, organization, tenant, identityUri, debug, insecure)
		return err
	})
	return response, err
}

func (c DataServiceCommand) downloadEntitySpecification(baseUri url.URL, organization string, tenant string, identityUri url.URL, debug bool, insecure bool) ([]byte, error) {
	uri := c.formatUri(baseUri, organization, tenant) + "/DataService"
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}

	config := c.configProvider.Config("default")
	auth, err := c.executeAuthenticators(config.Auth, identityUri, debug, insecure, request)
	if err != nil {
		return nil, err
	}
	for key, value := range auth.RequestHeader {
		request.Header.Add(key, value)
	}
	response, err := c.sendRequest(request, insecure)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, resiliency.Retryable(fmt.Errorf("Error reading response: %w", err))
	}
	if response.StatusCode >= 500 {
		return nil, resiliency.Retryable(fmt.Errorf("Dataservice returned status code '%v' and body '%v'", response.StatusCode, string(body)))
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Dataservice returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return body, nil
}

func (c DataServiceCommand) sendRequest(request *http.Request, insecure bool) (*http.Response, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint // This is user configurable and disabled by default
	}
	client := &http.Client{Transport: transport}
	return client.Do(request)
}

func (c DataServiceCommand) executeAuthenticators(authConfig config.AuthConfig, identityUri url.URL, debug bool, insecure bool, request *http.Request) (*auth.AuthenticatorResult, error) {
	authRequest := *auth.NewAuthenticatorRequest(request.URL.String(), map[string]string{})
	ctx := *auth.NewAuthenticatorContext(authConfig.Type, authConfig.Config, identityUri, debug, insecure, authRequest)
	for _, authProvider := range c.authenticators {
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
		path = "/{organization}/{tenant}/dataservice_/api"
	}
	path = strings.ReplaceAll(path, "{organization}", org)
	path = strings.ReplaceAll(path, "{tenant}", tenant)
	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("%s://%s%s", baseUri.Scheme, baseUri.Host, path)
}

func (c DataServiceCommand) logRequest(logger log.Logger, request *http.Request) {
	buffer := &bytes.Buffer{}
	_, _ = buffer.ReadFrom(request.Body)
	body := buffer.Bytes()
	request.Body = io.NopCloser(bytes.NewReader(body))
	requestInfo := log.NewRequestInfo(request.Method, request.URL.String(), request.Proto, request.Header, bytes.NewReader(body))
	logger.LogRequest(*requestInfo)
}

func (c DataServiceCommand) logResponse(logger log.Logger, response *http.Response, body []byte) {
	responseInfo := log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body))
	logger.LogResponse(*responseInfo)
}

func NewDataServiceCommand(args []string, stdErr io.Writer, configProvider config.ConfigProvider, authenticators []auth.Authenticator) *DataServiceCommand {
	return &DataServiceCommand{args, stdErr, configProvider, authenticators}
}
