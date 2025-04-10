//go:build !windows

package studio

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/test"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestPackOnLinuxWithCorrectArguments(t *testing.T) {
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

	if !strings.HasSuffix(commandName, "dotnet") {
		t.Errorf("Expected command name to be dotnet, but got: %v", commandName)
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
	if output != destination {
		t.Errorf("Expected --output argument to be %s, but got: %v", destination, commandArgs)
	}
}

func TestAnalyzeOnLinuxWithCorrectArguments(t *testing.T) {
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

	if !strings.HasSuffix(commandName, "dotnet") {
		t.Errorf("Expected command name to be dotnet, but got: %v", commandName)
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

func TestAnalyzeWindowsProjectOnLinuxReturnsCompatibilityError(t *testing.T) {
	called := false
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		called = true
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := studioWindowsProjectDirectory()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	if called {
		t.Error("Expected uipcli not to be called but it was.")
	}
	if result.Error == nil || result.Error.Error() != "UiPath Studio Projects which target windows-only are not support on linux devices. Build the project on windows or change the target framework to cross-platform." {
		t.Errorf("Expected compatibility error, but got: %v", result.Error)
	}
}

func TestPackWindowsProjectOnLinuxReturnsCompatibilityError(t *testing.T) {
	called := false
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		called = true
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := studioWindowsProjectDirectory()
	destination := createDirectory(t)
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	if called {
		t.Error("Expected uipcli not to be called but it was.")
	}
	if result.Error == nil || result.Error.Error() != "UiPath Studio Projects which target windows-only are not support on linux devices. Build the project on windows or change the target framework to cross-platform." {
		t.Errorf("Expected compatibility error, but got: %v", result.Error)
	}
}

func TestRunOnLinuxWithCorrectPackArguments(t *testing.T) {
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

	if !strings.HasSuffix(commandName, "dotnet") {
		t.Errorf("Expected command name to be dotnet, but got: %v", commandName)
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
