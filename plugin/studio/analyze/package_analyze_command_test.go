package analyze

import (
	"reflect"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestAnalyzeNonExistentProjectShowsProjectJsonNotFound(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "analyze", "--source", "non-existent"}, context)

	if !strings.Contains(result.StdErr, "project.json not found") {
		t.Errorf("Expected stderr to show that project.json was not found, but got: %v", result.StdErr)
	}
}

func TestAnalyzeCrossPlatformSuccessfully(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := test.NewCrossPlatformProject(t).
		WithDefaultGovernanceFile().
		Build()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	if result.Error != nil {
		t.Errorf("Expected error to be nil, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status to be Succeeded, but got: %v", result.StdOut)
	}
	violations := stdout["violations"].([]interface{})
	if len(violations) == 0 {
		t.Errorf("Expected violations not to be empty, but got: %v", result.StdOut)
	}

	expected := map[string]interface{}{
		"activityDisplayName": "",
		"activityId":          nil,
		"description":         "Workflow Main.xaml does not have any assigned Test Cases.",
		"documentationLink":   "https://docs.uipath.com/activities/lang-en/docs/ta-dbp-002",
		"errorCode":           "TA-DBP-002",
		"errorSeverity":       2.0,
		"severity":            "Warning",
		"filePath":            "",
		"item":                nil,
		"recommendation":      "Creating Test Cases for your workflows allows you to run them frequently to discover potential issues early on before they are introduced in your production environment. [Learn more.](https://docs.uipath.com/activities/lang-en/docs/ta-dbp-002)",
		"ruleName":            "Untested Workflows",
		"workflowDisplayName": "Main",
	}
	violation := findViolation(violations, "TA-DBP-002")
	if !reflect.DeepEqual(expected, violation) {
		t.Errorf("Expected violation '%v', but got: '%v'", expected, violation)
	}
}

func TestAnalyzeCrossPlatformWithTreatWarningAsErrors(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source, "--treat-warnings-as-errors", "true"}, context)

	if result.Error == nil {
		t.Errorf("Expected error, but got nil")
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Failed" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
	if stdout["error"] != nil {
		t.Errorf("Expected no error message, but got: %v", result.StdOut)
	}
}

func TestAnalyzeReturnsErrorStatus(t *testing.T) {
	exec := process.NewExecCustomProcess(1, "Analyze output", "There was an error", func(name string, args []string) {})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Error" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
	if stdout["error"] != "There was an error" {
		t.Errorf("Expected error to be set, but got: %v", result.StdOut)
	}
}

func TestAnalyzeCrossPlatformWithGovernanceFileSuccessfully(t *testing.T) {
	governanceFile := test.CreateFileWithContent(t, `
{
  "product-name": "Development",
  "policy-name": "Modern Policy - Development",
  "data": {
    "embedded-rules-config-rules": [
      {
        "code-embedded-rules-config-rules": "ST-USG-010",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Warning",
        "parameters-embedded-rules-config-rules": []
      }
	]
  }
}
`)

	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := test.NewCrossPlatformProject(t).
		WithDefaultGovernanceFile().
		Build()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source, "--governance-file", governanceFile}, context)

	if result.Error != nil {
		t.Errorf("Expected error to be nil, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status to be Succeeded, but got: %v", result.StdOut)
	}
	if stdout["error"] != nil {
		t.Errorf("Expected no error message, but got: %v", result.StdOut)
	}
}

func TestAnalyzeCrossPlatformWithGovernanceFileViolations(t *testing.T) {
	governanceFile := test.CreateFileWithContent(t, `
{
  "product-name": "Development",
  "policy-name": "Modern Policy - Development",
  "data": {
    "embedded-rules-config-rules": [
      {
        "code-embedded-rules-config-rules": "ST-USG-010",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      }
	]
  }
}
`)

	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source, "--governance-file", governanceFile}, context)

	if result.Error == nil {
		t.Errorf("Expected error not to be nil, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Failed" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
	violations := stdout["violations"].([]interface{})
	if len(violations) == 0 {
		t.Errorf("Expected violations not to be empty, but got: %v", result.StdOut)
	}

	expected := map[string]interface{}{
		"activityDisplayName": "",
		"activityId":          nil,
		"description":         "Dependency package UiPath.Testing.Activities is not used.",
		"documentationLink":   "https://docs.uipath.com/studio/lang-en/2024.10/docs/st-usg-010",
		"errorCode":           "ST-USG-010",
		"errorSeverity":       1.0,
		"severity":            "Error",
		"filePath":            "",
		"item": map[string]interface{}{
			"name": "UiPath.Testing.Activities",
			"type": 4.0,
		},
		"recommendation":      "Remove unused packages in order to improve process execution time. [Learn more.](https://docs.uipath.com/studio/lang-en/2024.10/docs/st-usg-010)",
		"ruleName":            "Unused Dependencies",
		"workflowDisplayName": "",
	}
	violation := findViolation(violations, "ST-USG-010")
	if !reflect.DeepEqual(expected, violation) {
		t.Errorf("Expected violation '%v', but got: '%v'", expected, violation)
	}
}

func TestAnalyzeGovernanceFileViolationsWithoutStopOnRuleViolationReturnsNoError(t *testing.T) {
	governanceFile := test.CreateFileWithContent(t, `
{
  "product-name": "Development",
  "policy-name": "Modern Policy - Development",
  "data": {
    "embedded-rules-config-rules": [
      {
        "code-embedded-rules-config-rules": "ST-USG-010",
        "is-enabled-embedded-rules-config-rules": true,
        "default-action": "Error",
        "parameters-embedded-rules-config-rules": []
      }
	]
  }
}
`)

	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source, "--governance-file", governanceFile, "--stop-on-rule-violation", "false"}, context)

	if result.Error != nil {
		t.Errorf("Expected error to be nil, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Failed" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
}

func TestAnalyzeUnknownGovernanceReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source, "--governance-file", "unknown-governance-file"}, context)

	if result.Error == nil || result.Error.Error() != "unknown-governance-file not found" {
		t.Errorf("Expected governance file not found error, but got: %v", result.Error)
	}
}

func TestAnalyzeWithLibraryAuthentication(t *testing.T) {
	config := `
profiles:
  - name: default
    organization: my-org
    tenant: my-tenant
    auth:
      clientId: success-client-id
      clientSecret: success-client-secret
`
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(200, `{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	identityUrl := test.GetArgumentValue(commandArgs, "--identityUrl")
	if identityUrl != result.BaseUrl+"/identity_" {
		t.Errorf("Expected identity url as argument, but got: %v", commandArgs)
	}

	orchestratorUrl := test.GetArgumentValue(commandArgs, "--orchestratorUrl")
	if orchestratorUrl != result.BaseUrl {
		t.Errorf("Expected orchestrator url as argument, but got: %v", commandArgs)
	}

	authToken := test.GetArgumentValue(commandArgs, "--orchestratorAuthToken")
	if authToken != "my-jwt-access-token" {
		t.Errorf("Expected jwt bearer token as argument, but got: %v", commandArgs)
	}

	organization := test.GetArgumentValue(commandArgs, "--orchestratorAccountName")
	if organization != "my-org" {
		t.Errorf("Expected organization as argument, but got: %v", commandArgs)
	}

	tenant := test.GetArgumentValue(commandArgs, "--orchestratorTenant")
	if tenant != "my-tenant" {
		t.Errorf("Expected tenant as argument, but got: %v", commandArgs)
	}
}

func findViolation(violations []interface{}, errorCode string) map[string]interface{} {
	var violation map[string]interface{}
	for _, v := range violations {
		vMap := v.(map[string]interface{})
		if vMap["errorCode"] == errorCode {
			violation = vMap
		}
	}
	return violation
}
