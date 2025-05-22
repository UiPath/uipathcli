package test

import (
	"bytes"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
)

func TestPluginHidesOperation(t *testing.T) {
	definition := `
paths:
  /hidden:
    get:
      summary: This command should not be shown
      operationId: my-hidden-command
`

	context := NewContextBuilder().
		WithDefinition("mypluginservice", definition).
		WithCommandPlugin(HideOperationPluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "--help"}, context)

	if strings.Contains(result.StdOut, "my-hidden-command") {
		t.Errorf("Expected hidden command to not show up, but got: %v", result.StdOut)
	}
}

func TestPluginAddsNewOperation(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("mypluginservice", definition).
		WithCommandPlugin(SimplePluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "--help"}, context)

	if !strings.Contains(result.StdOut, "ping") {
		t.Errorf("Expected ping command to show up, but got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "my-plugin-command") {
		t.Errorf("Expected ping command to show up, but got: %v", result.StdOut)
	}
}

func TestPluginOverridesExistingOperation(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: This should not be shown
      operationId: my-plugin-command
`

	context := NewContextBuilder().
		WithDefinition("mypluginservice", definition).
		WithCommandPlugin(SimplePluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "--help"}, context)

	if !strings.Contains(result.StdOut, "This is a simple plugin command") {
		t.Errorf("Expected plugin command to show up, but got: %v", result.StdOut)
	}
}

func TestPluginInvokesOperation(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: This should not be shown
      operationId: my-plugin-command
`

	context := NewContextBuilder().
		WithDefinition("mypluginservice", definition).
		WithCommandPlugin(SimplePluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "my-plugin-command"}, context)

	if result.StdOut != "Simple plugin output" {
		t.Errorf("Expected plugin command to show response data, but got: %v", result.StdOut)
	}
	if result.StdErr != "Simple plugin logging output" {
		t.Errorf("Expected plugin command to show error log on stderr, but got: %v", result.StdErr)
	}
}

func TestPluginShowsDebugOutput(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: This should not be shown
      operationId: my-plugin-command
`

	context := NewContextBuilder().
		WithDefinition("mypluginservice", definition).
		WithCommandPlugin(SimplePluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "my-plugin-command", "--debug"}, context)

	if result.StdOut != "Simple plugin output" {
		t.Errorf("Expected plugin command to show response data, but got: %v", result.StdOut)
	}
	if result.StdErr != "Simple plugin logging output" {
		t.Errorf("Expected plugin command to show error log on stderr, but got: %v", result.StdErr)
	}
}

func TestShowsErrorFromPlugin(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: This should not be shown
      operationId: my-failed-command
`

	context := NewContextBuilder().
		WithDefinition("mypluginservice", definition).
		WithCommandPlugin(ErrorPluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "my-failed-command"}, context)

	if !strings.Contains(result.StdErr, "Internal server error when calling mypluginservice") {
		t.Errorf("Expected error from plugin command, but got: %v", result.StdErr)
	}
}

func TestPluginContextData(t *testing.T) {
	config := `
profiles:
- name: default
  auth:
    clientId: very
    clientSecret: short
`
	definition := `
paths:
  /ping:
    get:
      summary: This should not be shown
      operationId: my-plugin-command
`
	pluginCommand := ContextPluginCommand{}
	context := NewContextBuilder().
		WithDefinition("mypluginservice", definition).
		WithConfig(config).
		WithCommandPlugin(&pluginCommand).
		WithTokenResponse(http.StatusOK, `{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		Build()

	RunCli([]string{"mypluginservice", "my-plugin-command"}, context)

	if !strings.Contains(pluginCommand.Context.BaseUri.String(), "http://127.0.0.1") {
		t.Errorf("Expected plugin command to retrieve base uri, but got: %v", pluginCommand.Context.BaseUri.String())
	}
	expectedAuthTokenType := "Bearer"
	actualAuthTokenType := pluginCommand.Context.Auth.Token.Type
	if actualAuthTokenType != expectedAuthTokenType {
		t.Errorf("Expected plugin command to retrieve auth token type %v, but got: %v", expectedAuthTokenType, actualAuthTokenType)
	}
	expectedAuthToken := "my-jwt-access-token"
	actualAuthToken := pluginCommand.Context.Auth.Token.Value
	if actualAuthToken != expectedAuthToken {
		t.Errorf("Expected plugin command to retrieve auth token %v, but got: %v", expectedAuthToken, actualAuthToken)
	}
}

func TestPluginContextInsecureFlag(t *testing.T) {
	config := `
profiles:
- name: default
  insecure: true
  auth:
    clientId: very
    clientSecret: short
`
	pluginCommand := ContextPluginCommand{}
	context := NewContextBuilder().
		WithDefinition("mypluginservice", "").
		WithConfig(config).
		WithCommandPlugin(&pluginCommand).
		WithTokenResponse(http.StatusOK, `{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		Build()

	RunCli([]string{"mypluginservice", "my-plugin-command"}, context)

	if !pluginCommand.Context.Settings.Insecure {
		t.Errorf("Expected insecure flag to be true, but got: %v", pluginCommand.Context.Settings.Insecure)
	}
}

func TestPluginContextParameterValue(t *testing.T) {
	pluginCommand := ContextPluginCommand{}
	context := NewContextBuilder().
		WithDefinition("mypluginservice", "").
		WithCommandPlugin(&pluginCommand).
		Build()

	RunCli([]string{"mypluginservice", "my-plugin-command", "--filter", "my-filter"}, context)

	numberOfParameters := len(pluginCommand.Context.Parameters)
	if numberOfParameters != 1 {
		t.Errorf("Expected one parameter, but got: %v", numberOfParameters)
	}
	name := pluginCommand.Context.Parameters[0].Name
	if name != "filter" {
		t.Errorf("Expected 'filter' parameter, but got: %v", name)
	}
	value := pluginCommand.Context.Parameters[0].Value
	if value != "my-filter" {
		t.Errorf("Expected parameter value %v, but got: %v", "my-filter", value)
	}
}

func TestPluginShowsParameter(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("mypluginservice", "").
		WithCommandPlugin(ParametrizedPluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "my-parametrized-command", "--help"}, context)

	expectedName := "--take"
	if !strings.Contains(result.StdOut, expectedName) {
		t.Errorf("Expected parameter %v to be shown, but got: %v", expectedName, result.StdOut)
	}
	expectedDescription := "This is a take parameter"
	if !strings.Contains(result.StdOut, expectedDescription) {
		t.Errorf("Expected parameter description '%v' to be shown, but got: %v", expectedDescription, result.StdOut)
	}
}

func TestPluginRequiresParameter(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("mypluginservice", "").
		WithCommandPlugin(ParametrizedPluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "my-parametrized-command"}, context)

	expected := "Argument --take is missing"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("Expected error message for missing parameter %v, but got: %v", expected, result.StdErr)
	}
}

func TestPluginValidatesParameterType(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("mypluginservice", "").
		WithCommandPlugin(ParametrizedPluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "my-parametrized-command", "--take", "test"}, context)

	expected := "Cannot convert 'take' value 'test' to integer"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("Expected error message for invalid parameter value %v, but got: %v", expected, result.StdErr)
	}
}

func TestPluginShowsParameterDefaultValue(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("mypluginservice", "").
		WithCommandPlugin(ParametrizedPluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "my-parametrized-command", "--help"}, context)

	if !strings.Contains(result.StdOut, "--filter string (default: all)") {
		t.Errorf("Expected default value in help output, but got: %v", result.StdOut)
	}
}

func TestPluginShowsParameterAllowedValues(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("mypluginservice", "").
		WithCommandPlugin(ParametrizedPluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "my-parametrized-command", "--help"}, context)

	if !strings.Contains(result.StdOut, "Allowed values:") {
		t.Errorf("stdout does not contain allowed values, got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "- all") {
		t.Errorf("stdout does not contain first allowed value, got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "- default") {
		t.Errorf("stdout does not contain second allowed value, got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "- none") {
		t.Errorf("stdout does not third second allowed value, got: %v", result.StdOut)
	}
}

func TestPluginDoesNotShowHiddenParameter(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("mypluginservice", "").
		WithCommandPlugin(ParametrizedPluginCommand{}).
		Build()

	result := RunCli([]string{"mypluginservice", "my-parametrized-command", "--help"}, context)

	if strings.Contains(result.StdOut, "--skip") {
		t.Errorf("Expected help output not to show hidden parameter, but got: %v", result.StdOut)
	}
}

type SimplePluginCommand struct{}

func (c SimplePluginCommand) Command() plugin.Command {
	return *plugin.NewCommand("mypluginservice").
		WithOperation("my-plugin-command", "Simple Command", "This is a simple plugin command")
}

func (c SimplePluginCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	logger.LogError("Simple plugin logging output")
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "https", map[string][]string{}, bytes.NewReader([]byte("Simple plugin output"))))
}

type ContextPluginCommand struct {
	Context plugin.ExecutionContext
}

func (c *ContextPluginCommand) Command() plugin.Command {
	return *plugin.NewCommand("mypluginservice").
		WithOperation("my-plugin-command", "Simple Command", "This is a simple plugin command").
		WithParameter(plugin.NewParameter("filter", plugin.ParameterTypeString, "This is a filter"))
}

func (c *ContextPluginCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	c.Context = ctx
	return nil
}

type ErrorPluginCommand struct{}

func (c ErrorPluginCommand) Command() plugin.Command {
	return *plugin.NewCommand("mypluginservice").
		WithOperation("my-failed-command", "Command fails", "This command always fails")
}

func (c ErrorPluginCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	return errors.New("Internal server error when calling mypluginservice")
}

type HideOperationPluginCommand struct{}

func (c HideOperationPluginCommand) Command() plugin.Command {
	return *plugin.NewCommand("mypluginservice").
		WithOperation("my-hidden-command", "Hidden command", "This command should not be shown").
		IsHidden()
}

func (c HideOperationPluginCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	return errors.New("my-hidden-command is not supported")
}

type ParametrizedPluginCommand struct{}

func (c ParametrizedPluginCommand) Command() plugin.Command {
	return *plugin.NewCommand("mypluginservice").
		WithOperation("my-parametrized-command", "Parametrized Command", "This is a plugin command with parameters").
		WithParameter(plugin.NewParameter("skip", plugin.ParameterTypeInteger, "This is a skip parameter").
			WithHidden(true)).
		WithParameter(plugin.NewParameter("take", plugin.ParameterTypeInteger, "This is a take parameter").
			WithRequired(true)).
		WithParameter(plugin.NewParameter("filter", plugin.ParameterTypeString, "This is a filter parameter").
			WithDefaultValue("all").
			WithAllowedValues([]interface{}{"all", "default", "none"}))
}

func (c ParametrizedPluginCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	return nil
}
