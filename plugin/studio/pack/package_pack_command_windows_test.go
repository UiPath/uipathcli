//go:build windows

package pack

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestPackWindowsSuccessfully(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()

	source := test.NewWindowsProject(t).
		Build()
	destination := test.CreateDirectory(t)
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	stdout := test.ParseOutput(t, result.StdOut)
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
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	destination := test.CreateDirectory(t)
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
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := test.NewWindowsProject(t).
		Build()
	destination := test.CreateDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	if !strings.HasSuffix(commandName, "uipcli.exe") {
		t.Errorf("Expected command name to be uipcli.exe, but got: %v", commandName)
	}

	expectedArgs := []string{"package", "pack", filepath.Join(source, "project.json"), "--output", destination}
	if !reflect.DeepEqual(expectedArgs, commandArgs) {
		t.Errorf("Expected arguments to be '%v', but got: '%v'", expectedArgs, commandArgs)
	}
}
