//go:build windows

package studio

import (
	"encoding/json"
	"fmt"
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

func TestParallelRunPassed(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewTestRunCommand()).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyProcess_Tests'", 200, `{"value":[{"id":10000,"name":"MyProcess_Tests"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyWindowsProcess_Tests'", 200, `{"value":[{"id":20000,"name":"MyWindowsProcess_Tests"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases(10000)", 200, `{"name":"MyProcess_Tests","processKey":"MyProcess_Tests","processVersion":"1.0.0"}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases(20000)", 200, `{"name":"MyWindowsProcess_Tests","processKey":"MyWindowsProcess_Tests","processVersion":"2.0.0"}`).
		WithResponseHandler(func(request test.RequestData) test.ResponseData {
			if request.URL.Path == "/my-org/my-tenant/orchestrator_/api/TestAutomation/CreateTestSetForReleaseVersion" {
				body := map[string]interface{}{}
				err := json.Unmarshal(request.Body, &body)
				if err != nil {
					return test.ResponseData{Status: 500, Body: err.Error()}
				}
				if body["releaseId"] == 10000.0 {
					return test.ResponseData{Status: 201, Body: "100002"}
				}
				return test.ResponseData{Status: 201, Body: "200002"}
			}
			return test.ResponseData{Status: 500, Body: fmt.Sprintf("Unhandled HTTP request %s", request.URL.Path)}
		}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=100002&triggerType=ExternalTool", 200, "100001").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=200002&triggerType=ExternalTool", 200, "200001").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSetExecutions(100001)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", 200,
			`{
               "Name":"Automated - MyProcess_Tests - 1.0.0",
               "TestSetId":100002,
               "StartTime":"2025-03-17T12:10:09.053Z",
               "EndTime":"2025-03-17T12:10:18.183Z",
               "Status":"Passed",
               "Id":100001,
               "TestCaseExecutions":[{
                 "Id":100003,
                 "TestCaseId":100004,
                 "EntryPointPath":"TestCase.xaml",
                 "StartTime":"2025-03-17T12:10:09.087Z",
                 "EndTime":"2025-03-17T12:10:18.083Z",
                 "Status":"Passed"
               }]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSetExecutions(200001)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", 200,
			`{
               "Name":"Automated - MyWindowsProcess_Tests - 2.0.0",
               "TestSetId":200002,
               "StartTime":"2025-03-17T12:10:09.053Z",
               "EndTime":"2025-03-17T12:10:18.183Z",
               "Status":"Failed",
               "Id":200001,
               "TestCaseExecutions":[{
                 "Id":200003,
                 "TestCaseId":200004,
                 "EntryPointPath":"TestCase.xaml",
                 "StartTime":"2025-03-17T12:10:09.087Z",
                 "EndTime":"2025-03-17T12:10:18.083Z",
                 "Status":"Failed"
               }]
             }`).
		Build()

	source1 := studioCrossPlatformProjectDirectory()
	source2 := studioWindowsProjectDirectory()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source1 + "," + source2, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	stdout := parseOutput(t, result.StdOut)
	expected := map[string]interface{}{
		"testSetExecutions": []interface{}{
			map[string]interface{}{
				"canceledCount": 0.0,
				"endTime":       "2025-03-17T12:10:18.183Z",
				"failuresCount": 0.0,
				"id":            100001.0,
				"name":          "Automated - MyProcess_Tests - 1.0.0",
				"passedCount":   1.0,
				"startTime":     "2025-03-17T12:10:09.053Z",
				"status":        "Passed",
				"testCaseExecutions": []interface{}{
					map[string]interface{}{
						"endTime":    "2025-03-17T12:10:18.083Z",
						"error":      nil,
						"id":         100003.0,
						"name":       "TestCase.xaml",
						"startTime":  "2025-03-17T12:10:09.087Z",
						"status":     "Passed",
						"testCaseId": 100004.0,
					},
				},
				"testCasesCount": 1.0,
				"testSetId":      100002.0,
			},
			map[string]interface{}{
				"canceledCount": 0.0,
				"endTime":       "2025-03-17T12:10:18.183Z",
				"failuresCount": 1.0,
				"id":            200001.0,
				"name":          "Automated - MyWindowsProcess_Tests - 2.0.0",
				"passedCount":   0.0,
				"startTime":     "2025-03-17T12:10:09.053Z",
				"status":        "Failed",
				"testCaseExecutions": []interface{}{
					map[string]interface{}{
						"endTime":    "2025-03-17T12:10:18.083Z",
						"error":      nil,
						"id":         200003.0,
						"name":       "TestCase.xaml",
						"startTime":  "2025-03-17T12:10:09.087Z",
						"status":     "Failed",
						"testCaseId": 200004.0,
					},
				},
				"testCasesCount": 1.0,
				"testSetId":      200002.0,
			},
		},
	}
	if !reflect.DeepEqual(expected, stdout) {
		t.Errorf("Expected output '%v', but got: '%v'", expected, stdout)
	}
}
