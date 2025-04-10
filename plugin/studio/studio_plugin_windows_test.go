//go:build windows

package studio

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/test"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestPackWindowsSuccessfully(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()

	source := studioWindowsProjectDirectory()
	destination := createDirectory(t)
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	stdout := parseOutput(t, result.StdOut)
	outputFile := filepath.Join(destination, "MyWindowsProcess.1.0.0.nupkg")
	expected := map[string]interface{}{
		"status":      "Succeeded",
		"error":       nil,
		"name":        "MyWindowsProcess",
		"description": "Blank Process",
		"projectId":   "94c4c9c1-68c3-45d4-be14-d4427f17eddd",
		"version":     "1.0.0",
		"output":      outputFile,
	}
	if !reflect.DeepEqual(expected, stdout) {
		t.Errorf("Expected output '%v', but got: '%v'", expected, stdout)
	}
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("Expected output file %s to exists, but could not find it: %v", outputFile, err)
	}
}

func TestPackCrossPlatformProjectOnWindowsWithCorrectArguments(t *testing.T) {
	commandName := ""
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandName = name
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
	destination := createDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	if !strings.HasSuffix(commandName, "dotnet.exe") {
		t.Errorf("Expected command name to be dotnet.exe, but got: %v", commandName)
	}
	if !strings.HasSuffix(commandArgs[0], "uipcli.dll") {
		t.Errorf("Expected 1st argument to be the uipcli.dll, but got: %v", commandArgs[0])
	}

	expectedArgs := []string{"package", "pack", filepath.Join(source, "project.json"), "--output", destination}
	actualArgs := commandArgs[1:]
	if !reflect.DeepEqual(expectedArgs, actualArgs) {
		t.Errorf("Expected arguments to be '%v', but got: '%v'", expectedArgs, actualArgs)
	}
}

func TestPackWindowsOnlyProjectOnWindowsWithCorrectArguments(t *testing.T) {
	commandName := ""
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandName = name
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := studioWindowsProjectDirectory()
	destination := createDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	if !strings.HasSuffix(commandName, "uipcli.exe") {
		t.Errorf("Expected command name to be uipcli.exe, but got: %v", commandName)
	}

	expectedArgs := []string{"package", "pack", filepath.Join(source, "project.json"), "--output", destination}
	if !reflect.DeepEqual(expectedArgs, commandArgs) {
		t.Errorf("Expected arguments to be '%v', but got: '%v'", expectedArgs, commandArgs)
	}
}

func TestAnalyzeWindowsWithErrors(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := studioWindowsProjectDirectory()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	if result.Error == nil {
		t.Errorf("Expected error not to be nil, but got: %v", result.Error)
	}
	stdout := parseOutput(t, result.StdOut)
	if stdout["status"] != "Failed" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
	if stdout["error"] != nil {
		t.Errorf("Expected no standard error output, but got: %v", result.StdOut)
	}
	violations := stdout["violations"].([]interface{})
	if len(violations) == 0 {
		t.Errorf("Expected violations not to be empty, but got: %v", result.StdOut)
	}
}

func TestAnalyzeWindowsWithErrorsButStopOnRuleViolationFalse(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := studioWindowsProjectDirectory()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source, "--stop-on-rule-violation", "false"}, context)

	if result.Error != nil {
		t.Errorf("Expected error to be nil, but got: %v", result.Error)
	}
	stdout := parseOutput(t, result.StdOut)
	if stdout["status"] != "Failed" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
	if stdout["error"] != nil {
		t.Errorf("Expected no standard error output, but got: %v", result.StdOut)
	}
	violations := stdout["violations"].([]interface{})
	if len(violations) == 0 {
		t.Errorf("Expected violations not to be empty, but got: %v", result.StdOut)
	}
}

func TestAnalyzeWindowsOnlyProjectOnWindowsWithCorrectArguments(t *testing.T) {
	commandName := ""
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandName = name
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := studioWindowsProjectDirectory()
	test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	if !strings.HasSuffix(commandName, "uipcli.exe") {
		t.Errorf("Expected command name to be uipcli.exe, but got: %v", commandName)
	}
	if commandArgs[0] != "package" {
		t.Errorf("Expected 1st argument to be package, but got: %v", commandArgs[0])
	}
	if commandArgs[1] != "analyze" {
		t.Errorf("Expected 2nd argument to be analyze, but got: %v", commandArgs[1])
	}
	if commandArgs[2] != filepath.Join(source, "project.json") {
		t.Errorf("Expected 3rd argument to be the project.json, but got: %v", commandArgs[2])
	}
}

func TestAnalyzeCrossPlatformProjectOnWindowsWithCorrectArguments(t *testing.T) {
	commandName := ""
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandName = name
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
	test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	if !strings.HasSuffix(commandName, "dotnet.exe") {
		t.Errorf("Expected command name to be dotnet.exe, but got: %v", commandName)
	}
	if !strings.HasSuffix(commandArgs[0], "uipcli.dll") {
		t.Errorf("Expected 1st argument to be the uipcli.dll, but got: %v", commandArgs[0])
	}
	if commandArgs[1] != "package" {
		t.Errorf("Expected 2nd argument to be package, but got: %v", commandArgs[1])
	}
	if commandArgs[2] != "analyze" {
		t.Errorf("Expected 3rd argument to be analyze, but got: %v", commandArgs[2])
	}
	if commandArgs[3] != filepath.Join(source, "project.json") {
		t.Errorf("Expected 4th argument to be the project.json, but got: %v", commandArgs[3])
	}
}

func TestRunOnWindowsWithCorrectPackArguments(t *testing.T) {
	commandName := ""
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		commandName = name
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
	test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	if !strings.HasSuffix(commandName, "dotnet.exe") {
		t.Errorf("Expected command name to be dotnet.exe, but got: %v", commandName)
	}
	if !strings.HasSuffix(commandArgs[0], "uipcli.dll") {
		t.Errorf("Expected 1st argument to be the uipcli.dll, but got: %v", commandArgs[0])
	}
	if commandArgs[1] != "package" {
		t.Errorf("Expected 2nd argument to be package, but got: %v", commandArgs[1])
	}
	if commandArgs[2] != "pack" {
		t.Errorf("Expected 3rd argument to be pack, but got: %v", commandArgs[2])
	}
	if commandArgs[3] != filepath.Join(source, "project.json") {
		t.Errorf("Expected 4th argument to be the project.json, but got: %v", commandArgs[3])
	}
	output := getArgumentValue(commandArgs, "--output")
	if output == "" {
		t.Errorf("Expected --output argument to be set, but got: %v", commandArgs)
	}
	outputType := getArgumentValue(commandArgs, "--outputType")
	if outputType != "Tests" {
		t.Errorf("Expected --outputType argument to be Tests, but got: %v", commandArgs)
	}
	if !slices.Contains(commandArgs, "--autoVersion") {
		t.Errorf("Expected --autoVersion argument to be set, but got: %v", commandArgs)
	}
}

func TestRestoreWindowsSuccessfully(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackageRestoreCommand()).
		Build()

	source := studioWindowsProjectDirectory()
	destination := createDirectory(t)
	result := test.RunCli([]string{"studio", "package", "restore", "--source", source, "--destination", destination}, context)

	stdout := parseOutput(t, result.StdOut)
	expected := map[string]interface{}{
		"status":      "Succeeded",
		"error":       nil,
		"name":        "MyWindowsProcess",
		"description": "Blank Process",
		"projectId":   "94c4c9c1-68c3-45d4-be14-d4427f17eddd",
		"output":      destination,
	}
	if !reflect.DeepEqual(expected, stdout) {
		t.Errorf("Expected output '%v', but got: '%v'", expected, stdout)
	}
}
