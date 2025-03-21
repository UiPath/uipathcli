//go:build windows

package studio

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	stdout := map[string]interface{}{}
	err := json.Unmarshal([]byte(result.StdOut), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize pack command result: %v", err)
	}
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status to be Succeeded, but got: %v", result.StdOut)
	}
	if stdout["error"] != nil {
		t.Errorf("Expected error to be nil, but got: %v", result.StdOut)
	}
	if stdout["name"] != "MyWindowsProcess" {
		t.Errorf("Expected name to be set, but got: %v", result.StdOut)
	}
	if stdout["description"] != "Blank Process" {
		t.Errorf("Expected version to be set, but got: %v", result.StdOut)
	}
	if stdout["projectId"] != "94c4c9c1-68c3-45d4-be14-d4427f17eddd" {
		t.Errorf("Expected projectId to be set, but got: %v", result.StdOut)
	}
	if stdout["version"] != "1.0.0" {
		t.Errorf("Expected version to be set, but got: %v", result.StdOut)
	}
	outputFile := stdout["output"].(string)
	if outputFile != filepath.Join(destination, "MyWindowsProcess.1.0.0.nupkg") {
		t.Errorf("Expected output path to be set, but got: %v", result.StdOut)
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
	if commandArgs[1] != "package" {
		t.Errorf("Expected 2nd argument to be package, but got: %v", commandArgs[1])
	}
	if commandArgs[2] != "pack" {
		t.Errorf("Expected 3rd argument to be pack, but got: %v", commandArgs[2])
	}
	if commandArgs[3] != filepath.Join(source, "project.json") {
		t.Errorf("Expected 4th argument to be the project.json, but got: %v", commandArgs[3])
	}
	if commandArgs[4] != "--output" {
		t.Errorf("Expected 5th argument to be output, but got: %v", commandArgs[4])
	}
	if commandArgs[5] != destination {
		t.Errorf("Expected 6th argument to be the output path, but got: %v", commandArgs[5])
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
	if commandArgs[0] != "package" {
		t.Errorf("Expected 1st argument to be package, but got: %v", commandArgs[0])
	}
	if commandArgs[1] != "pack" {
		t.Errorf("Expected 2nd argument to be pack, but got: %v", commandArgs[1])
	}
	if commandArgs[2] != filepath.Join(source, "project.json") {
		t.Errorf("Expected 3rd argument to be the project.json, but got: %v", commandArgs[2])
	}
	if commandArgs[3] != "--output" {
		t.Errorf("Expected 4th argument to be output, but got: %v", commandArgs[3])
	}
	if commandArgs[4] != destination {
		t.Errorf("Expected 5th argument to be the output path, but got: %v", commandArgs[4])
	}
}

func TestAnalyzeWindowsWithErrors(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := studioWindowsProjectDirectory()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	stdout := map[string]interface{}{}
	err := json.Unmarshal([]byte(result.StdOut), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize analyze command result: %v", err)
	}
	if result.Error == nil {
		t.Errorf("Expected error not to be nil, but got: %v", result.Error)
	}
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

	stdout := map[string]interface{}{}
	err := json.Unmarshal([]byte(result.StdOut), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize analyze command result: %v", err)
	}
	if result.Error != nil {
		t.Errorf("Expected error to be nil, but got: %v", result.Error)
	}
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
	if commandArgs[4] != "--outputType" {
		t.Errorf("Expected 5th argument to be outputType, but got: %v", commandArgs[4])
	}
	if commandArgs[5] != "Tests" {
		t.Errorf("Expected 6th argument to be Tests, but got: %v", commandArgs[5])
	}
	if commandArgs[6] != "--autoVersion" {
		t.Errorf("Expected 7th argument to be autoVersion, but got: %v", commandArgs[6])
	}
	if commandArgs[7] != "--output" {
		t.Errorf("Expected 8th argument to be output, but got: %v", commandArgs[7])
	}
	if commandArgs[8] == "" {
		t.Errorf("Expected 9th argument to be the output file, but got: %v", commandArgs[8])
	}
}
