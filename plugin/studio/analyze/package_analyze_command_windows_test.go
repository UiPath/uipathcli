//go:build windows

package analyze

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestAnalyzeWindowsWithErrors(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := test.NewWindowsProject(t).
		Build()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	if result.Error == nil {
		t.Errorf("Expected error not to be nil, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
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
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := test.NewWindowsProject(t).
		Build()
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
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
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
