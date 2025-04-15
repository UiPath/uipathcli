package pack

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestPackNonExistentProjectShowsProjectJsonNotFound(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "pack", "--source", "non-existent", "--destination", "test.nupkg"}, context)

	if !strings.Contains(result.StdErr, "project.json not found") {
		t.Errorf("Expected stderr to show that project.json was not found, but got: %v", result.StdErr)
	}
}

func TestPackInvalidOutputTypeShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()
	source := test.NewCrossPlatformProject(t).
		Build()
	destination := test.CreateDirectory(t)
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination, "--output-type", "unknown"}, context)

	if !strings.Contains(result.StdErr, "Invalid output type 'unknown', allowed values: Process, Library, Tests, Objects") {
		t.Errorf("Expected stderr to show output type is invalid, but got: %v", result.StdErr)
	}
}

func TestPackFailedReturnsFailureStatus(t *testing.T) {
	exec := process.NewExecCustomProcess(1, "Build output", "There was an error", func(name string, args []string) {})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	destination := test.CreateDirectory(t)
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Failed" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
	if stdout["error"] != "There was an error" {
		t.Errorf("Expected error to be set, but got: %v", result.StdOut)
	}
}

func TestPackCrossPlatformSuccessfully(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	destination := test.CreateDirectory(t)
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status to be Succeeded, but got: %v", result.StdOut)
	}
	if stdout["error"] != nil {
		t.Errorf("Expected error to be nil, but got: %v", result.StdOut)
	}
	if stdout["name"] != "MyProcess" {
		t.Errorf("Expected name to be set, but got: %v", result.StdOut)
	}
	if stdout["description"] != "Blank Process" {
		t.Errorf("Expected version to be set, but got: %v", result.StdOut)
	}
	if stdout["projectId"] != "9011ee47-8dd4-4726-8850-299bd6ef057c" {
		t.Errorf("Expected projectId to be set, but got: %v", result.StdOut)
	}
	if stdout["version"] != "1.0.0" {
		t.Errorf("Expected version to be set, but got: %v", result.StdOut)
	}
	outputFile := stdout["output"].(string)
	if outputFile != filepath.Join(destination, "MyProcess.1.0.0.nupkg") {
		t.Errorf("Expected output path to be set, but got: %v", result.StdOut)
	}
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("Expected output file %s to exists, but could not find it: %v", outputFile, err)
	}
}

func TestPackWithAutoVersionArgument(t *testing.T) {
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	destination := test.CreateDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination, "--auto-version", "true"}, context)

	if !slices.Contains(commandArgs, "--autoVersion") {
		t.Errorf("Expected --autoVersion argument to be set, but got: %v", commandArgs)
	}
}

func TestPackWithOutputTypeArgument(t *testing.T) {
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	destination := test.CreateDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination, "--output-type", "Process"}, context)

	outputType := test.GetArgumentValue(commandArgs, "--outputType")
	if outputType != "Process" {
		t.Errorf("Expected argument --outputType to be Process, but got: %v", commandArgs)
	}
}

func TestPackWithSplitOutputArgument(t *testing.T) {
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	destination := test.CreateDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination, "--split-output", "true"}, context)

	if !slices.Contains(commandArgs, "--splitOutput") {
		t.Errorf("Expected --splitOutput argument to be set, but got: %v", commandArgs)
	}
}

func TestPackWithReleaseNotesArgument(t *testing.T) {
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	destination := test.CreateDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination, "--release-notes", "These are release notes."}, context)

	releaseNotes := test.GetArgumentValue(commandArgs, "--releaseNotes")
	if releaseNotes != "These are release notes." {
		t.Errorf("Expected release notes argument, but got: %v", commandArgs)
	}
}

func TestPackWithLibraryAuthentication(t *testing.T) {
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
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	destination := test.CreateDirectory(t)
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	identityUrl := test.GetArgumentValue(commandArgs, "--libraryIdentityUrl")
	if identityUrl != result.BaseUrl+"/identity_" {
		t.Errorf("Expected identity url as argument, but got: %v", commandArgs)
	}

	orchestratorUrl := test.GetArgumentValue(commandArgs, "--libraryOrchestratorUrl")
	if orchestratorUrl != result.BaseUrl {
		t.Errorf("Expected orchestrator url as argument, but got: %v", commandArgs)
	}

	authToken := test.GetArgumentValue(commandArgs, "--libraryOrchestratorAuthToken")
	if authToken != "my-jwt-access-token" {
		t.Errorf("Expected jwt bearer token as argument, but got: %v", commandArgs)
	}

	organization := test.GetArgumentValue(commandArgs, "--libraryOrchestratorAccountName")
	if organization != "my-org" {
		t.Errorf("Expected organization as argument, but got: %v", commandArgs)
	}

	tenant := test.GetArgumentValue(commandArgs, "--libraryOrchestratorTenant")
	if tenant != "my-tenant" {
		t.Errorf("Expected tenant as argument, but got: %v", commandArgs)
	}
}

func TestPackWithOrgOnlyLibraryAuthentication(t *testing.T) {
	config := `
profiles:
  - name: default
    organization: my-org
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
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	destination := test.CreateDirectory(t)
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	identityUrl := test.GetArgumentValue(commandArgs, "--libraryIdentityUrl")
	if identityUrl != result.BaseUrl+"/identity_" {
		t.Errorf("Expected identity url as argument, but got: %v", commandArgs)
	}

	orchestratorUrl := test.GetArgumentValue(commandArgs, "--libraryOrchestratorUrl")
	if orchestratorUrl != result.BaseUrl {
		t.Errorf("Expected orchestrator url as argument, but got: %v", commandArgs)
	}

	authToken := test.GetArgumentValue(commandArgs, "--libraryOrchestratorAuthToken")
	if authToken != "my-jwt-access-token" {
		t.Errorf("Expected jwt bearer token as argument, but got: %v", commandArgs)
	}

	organization := test.GetArgumentValue(commandArgs, "--libraryOrchestratorAccountName")
	if organization != "my-org" {
		t.Errorf("Expected organization as argument, but got: %v", commandArgs)
	}

	tenant := test.GetArgumentValue(commandArgs, "--libraryOrchestratorTenant")
	if tenant != "" {
		t.Errorf("Expected no tenant as argument, but got: %v", commandArgs)
	}
}
