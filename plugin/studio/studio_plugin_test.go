package studio

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/test"
)

const studioDefinition = `
openapi: 3.0.1
info:
  title: UiPath Studio
  description: UiPath Studio
  version: v1
servers:
  - url: https://cloud.uipath.com/{organization}/studio_/backend
    description: The production url
    variables:
      organization:
        description: The organization name (or id)
        default: my-org
paths:
  {}
`

func TestPackWithoutSourceParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "pack", "--destination", "test.nupkg"}, context)

	if !strings.Contains(result.StdErr, "Argument --source is missing") {
		t.Errorf("Expected stderr to show that source parameter is missing, but got: %v", result.StdErr)
	}
}

func TestPackWithoutDestinationParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()

	source := studioProjectDirectory()
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source}, context)

	if !strings.Contains(result.StdErr, "Argument --destination is missing") {
		t.Errorf("Expected stderr to show that destination parameter is missing, but got: %v", result.StdErr)
	}
}

func TestPackNonExistentProject(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "pack", "--source", "non-existent", "--destination", "test.nupkg"}, context)

	if !strings.Contains(result.StdErr, "project.json not found") {
		t.Errorf("Expected stderr to show that project.json was not found, but got: %v", result.StdErr)
	}
}

func TestFailedPackagingReturnsFailureStatus(t *testing.T) {
	exec := test.NewFakeExecProcess(1, "Build output", "There was an error")
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := studioProjectDirectory()
	destination := createDirectory(t)
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination}, context)

	stdout := map[string]interface{}{}
	err := json.Unmarshal([]byte(result.StdOut), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize pack command result: %v", err)
	}
	if stdout["status"] != "Failed" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
	if stdout["error"] != "There was an error" {
		t.Errorf("Expected error to be set, but got: %v", result.StdOut)
	}
}

func TestPackSuccessfully(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()

	source := studioProjectDirectory()
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
	if stdout["name"] != "MyProcess" {
		t.Errorf("Expected name to be set, but got: %v", result.StdOut)
	}
	if stdout["description"] != "Blank Process" {
		t.Errorf("Expected version to be set, but got: %v", result.StdOut)
	}
	if stdout["projectId"] != "9011ee47-8dd4-4726-8850-299bd6ef057c" {
		t.Errorf("Expected projectId to be set, but got: %v", result.StdOut)
	}
	if stdout["version"] != "1.0.2" {
		t.Errorf("Expected version to be set, but got: %v", result.StdOut)
	}
	outputFile := stdout["output"].(string)
	if outputFile != filepath.Join(destination, "MyProcess.1.0.2.nupkg") {
		t.Errorf("Expected output path to be set, but got: %v", result.StdOut)
	}
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("Expected output file %s to exists, but could not find it: %v", outputFile, err)
	}
}

func TestAnalyzeWithoutSourceParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "analyze"}, context)

	if !strings.Contains(result.StdErr, "Argument --source is missing") {
		t.Errorf("Expected stderr to show that source parameter is missing, but got: %v", result.StdErr)
	}
}

func TestAnalyzeSuccessfully(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := studioProjectDirectory()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	stdout := map[string]interface{}{}
	err := json.Unmarshal([]byte(result.StdOut), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize analyze command result: %v", err)
	}
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status to be Succeeded, but got: %v", result.StdOut)
	}
	if stdout["error"] != nil {
		t.Errorf("Expected error to be nil, but got: %v", result.StdOut)
	}
	violations := stdout["violations"].([]interface{})
	if len(violations) == 0 {
		t.Errorf("Expected violations not to be empty, but got: %v", result.StdOut)
	}
	violation := findViolation(violations, "TA-DBP-002")
	if violation == nil {
		t.Errorf("Could not find violation TA-DBP-002, got: %v", result.StdOut)
	}
	if violation["activityDisplayName"] != "" {
		t.Errorf("Expected violation to have a activityDisplayName, but got: %v", result.StdOut)
	}
	if violation["description"] != "Workflow Main.xaml does not have any assigned Test Cases." {
		t.Errorf("Expected violation to have a description, but got: %v", result.StdOut)
	}
	if violation["documentationLink"] != "https://docs.uipath.com/activities/lang-en/docs/ta-dbp-002" {
		t.Errorf("Expected violation to have a documentationLink, but got: %v", result.StdOut)
	}
	if violation["errorSeverity"] != 1.0 {
		t.Errorf("Expected violation to have a errorSeverity, but got: %v", result.StdOut)
	}
	if violation["filePath"] != "" {
		t.Errorf("Expected violation to have a filePath, but got: %v", result.StdOut)
	}
	if violation["recommendation"] != "Creating Test Cases for your workflows allows you to run them frequently to discover potential issues early on before they are introduced in your production environment. [Learn more.](https://docs.uipath.com/activities/lang-en/docs/ta-dbp-002)" {
		t.Errorf("Expected violation to have a recommendation, but got: %v", result.StdOut)
	}
	if violation["ruleName"] != "Untested Workflows" {
		t.Errorf("Expected violation to have a ruleName, but got: %v", result.StdOut)
	}
	if violation["workflowDisplayName"] != "Main" {
		t.Errorf("Expected violation to have a workflowDisplayName, but got: %v", result.StdOut)
	}
}

func findViolation(violations []interface{}, errorCode string) map[string]interface{} {
	var violation map[string]interface{}
	for _, v := range violations {
		vMap := v.(map[string]interface{})
		if vMap["errorCode"] == errorCode {
			violation = vMap
		}
	}
	return violation
}

func studioProjectDirectory() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "project")
}

func createDirectory(t *testing.T) string {
	tmp, err := os.MkdirTemp("", "uipath-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmp) })
	return tmp
}
