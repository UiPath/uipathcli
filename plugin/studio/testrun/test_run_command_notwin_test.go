//go:build !windows

package testrun

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestRunOnLinuxWithCorrectPackArguments(t *testing.T) {
	commandName := ""
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		commandName = name
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
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
	output := test.GetArgumentValue(commandArgs, "--output")
	if output == "" {
		t.Errorf("Expected --output argument to be set, but got: %v", commandArgs)
	}
	outputType := test.GetArgumentValue(commandArgs, "--outputType")
	if outputType != "Tests" {
		t.Errorf("Expected --outputType argument to be Tests, but got: %v", commandArgs)
	}
	if !slices.Contains(commandArgs, "--autoVersion") {
		t.Errorf("Expected --autoVersion argument to be set, but got: %v", commandArgs)
	}
}
