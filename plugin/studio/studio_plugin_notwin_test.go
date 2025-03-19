//go:build !windows

package studio

import (
	"path/filepath"
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
	if commandArgs[4] != "--output" {
		t.Errorf("Expected 5th argument to be output, but got: %v", commandArgs[4])
	}
	if commandArgs[5] != destination {
		t.Errorf("Expected 6th argument to be the output path, but got: %v", commandArgs[5])
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
