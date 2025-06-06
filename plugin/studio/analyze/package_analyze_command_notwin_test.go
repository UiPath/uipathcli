//go:build !windows

package analyze

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestAnalyzeOnLinuxWithCorrectArguments(t *testing.T) {
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
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := test.NewWindowsProject(t).
		Build()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	if called {
		t.Error("Expected uipcli not to be called but it was.")
	}
	if result.Error == nil || result.Error.Error() != "UiPath Studio Projects which target windows-only are not support on linux devices. Build the project on windows or change the target framework to cross-platform." {
		t.Errorf("Expected compatibility error, but got: %v", result.Error)
	}
}
