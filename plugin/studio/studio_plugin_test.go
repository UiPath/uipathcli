package studio

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/UiPath/uipathcli/test"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestPackNonExistentProjectShowsProjectJsonNotFound(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "pack", "--source", "non-existent", "--destination", "test.nupkg"}, context)

	if !strings.Contains(result.StdErr, "project.json not found") {
		t.Errorf("Expected stderr to show that project.json was not found, but got: %v", result.StdErr)
	}
}

func TestInvalidOutputTypeShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()
	source := studioCrossPlatformProjectDirectory()
	destination := createDirectory(t)
	result := test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination, "--output-type", "unknown"}, context)

	if !strings.Contains(result.StdErr, "Invalid output type 'unknown', allowed values: Process, Library, Tests, Objects") {
		t.Errorf("Expected stderr to show output type is invalid, but got: %v", result.StdErr)
	}
}

func TestFailedPackagingReturnsFailureStatus(t *testing.T) {
	exec := process.NewExecCustomProcess(1, "Build output", "There was an error", func(name string, args []string) {})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
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

func TestPackCrossPlatformSuccessfully(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePackCommand()).
		Build()

	source := studioCrossPlatformProjectDirectory()
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
	if stdout["version"] != "1.0.0" {
		t.Errorf("Expected version to be set, but got: %v", result.StdOut)
	}
	outputFile := stdout["output"].(string)
	if outputFile != filepath.Join(destination, "MyProcess.1.0.0.nupkg") {
		t.Errorf("Expected output path to be set, but got: %v", result.StdOut)
	}
	if _, err := os.Stat(outputFile); err != nil {
		t.Errorf("Expected output file %s to exists, but could not find it: %v", outputFile, err)
	}
}

func TestPackWithAutoVersionArgument(t *testing.T) {
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
	destination := createDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination, "--auto-version", "true"}, context)

	if !slices.Contains(commandArgs, "--autoVersion") {
		t.Errorf("Expected argument --autoVersion, but got: %v", strings.Join(commandArgs, " "))
	}
}

func TestPackWithOutputTypeArgument(t *testing.T) {
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
	destination := createDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination, "--output-type", "Process"}, context)

	if !slices.Contains(commandArgs, "--outputType") {
		t.Errorf("Expected argument --outputType, but got: %v", strings.Join(commandArgs, " "))
	}
}

func TestPackWithSplitOutputArgument(t *testing.T) {
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
	destination := createDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination, "--split-output", "true"}, context)

	if !slices.Contains(commandArgs, "--splitOutput") {
		t.Errorf("Expected argument --splitOutput, but got: %v", strings.Join(commandArgs, " "))
	}
}

func TestPackWithReleaseNotesArgument(t *testing.T) {
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackagePackCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
	destination := createDirectory(t)
	test.RunCli([]string{"studio", "package", "pack", "--source", source, "--destination", destination, "--release-notes", "These are release notes."}, context)

	index := slices.Index(commandArgs, "--releaseNotes")
	if commandArgs[index] != "--releaseNotes" {
		t.Errorf("Expected argument --releaseNotes, but got: %v", strings.Join(commandArgs, " "))
	}
	if commandArgs[index+1] != "These are release notes." {
		t.Errorf("Expected release notes argument, but got: %v", strings.Join(commandArgs, " "))
	}
}

func TestAnalyzeNonExistentProjectShowsProjectJsonNotFound(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "analyze", "--source", "non-existent"}, context)

	if !strings.Contains(result.StdErr, "project.json not found") {
		t.Errorf("Expected stderr to show that project.json was not found, but got: %v", result.StdErr)
	}
}

func TestAnalyzeCrossPlatformSuccessfully(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackageAnalyzeCommand()).
		Build()

	source := studioCrossPlatformProjectDirectory()
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

func TestFailedAnalyzeReturnsFailureStatus(t *testing.T) {
	exec := process.NewExecCustomProcess(1, "Analyze output", "There was an error", func(name string, args []string) {})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
	result := test.RunCli([]string{"studio", "package", "analyze", "--source", source}, context)

	stdout := map[string]interface{}{}
	err := json.Unmarshal([]byte(result.StdOut), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize analyze command result: %v", err)
	}
	if stdout["status"] != "Failed" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
	if stdout["error"] != "There was an error" {
		t.Errorf("Expected error to be set, but got: %v", result.StdOut)
	}
}

func TestAnalyzeWithTreatWarningsAsErrorsArgument(t *testing.T) {
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
	test.RunCli([]string{"studio", "package", "analyze", "--source", source, "--treat-warnings-as-errors", "true"}, context)

	if !slices.Contains(commandArgs, "--treatWarningsAsErrors") {
		t.Errorf("Expected argument --treatWarningsAsErrors, but got: %v", strings.Join(commandArgs, " "))
	}
}

func TestAnalyzeWithStopOnRuleViolationArgument(t *testing.T) {
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(PackageAnalyzeCommand{exec}).
		Build()

	source := studioCrossPlatformProjectDirectory()
	test.RunCli([]string{"studio", "package", "analyze", "--source", source, "--stop-on-rule-violation", "true"}, context)

	if !slices.Contains(commandArgs, "--stopOnRuleViolation") {
		t.Errorf("Expected argument --stopOnRuleViolation, but got: %v", strings.Join(commandArgs, " "))
	}
}

func TestPublishNoPackageFileReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", "not-found"}, context)

	if result.Error == nil || result.Error.Error() != "Package not found." {
		t.Errorf("Expected package not found error, but got: %v", result.Error)
	}
}

func TestPublishMissingOrganizationReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--tenant", "my-tenant", "--source", "my.nupkg"}, context)

	if result.Error == nil || result.Error.Error() != "Organization is not set" {
		t.Errorf("Expected organization is not set error, but got: %v", result.Error)
	}
}

func TestPublishMissingTenantReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--source", "my.nupkg"}, context)

	if result.Error == nil || result.Error.Error() != "Tenant is not set" {
		t.Errorf("Expected tenant is not set error, but got: %v", result.Error)
	}
}

func TestPublishInvalidPackageReturnsError(t *testing.T) {
	path := createFile(t)
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", path}, context)

	if result.Error == nil || !strings.HasPrefix(result.Error.Error(), "Could not read package") {
		t.Errorf("Expected package read error, but got: %v", result.Error)
	}
}

func TestPublishReturnsPackageMetadata(t *testing.T) {
	nupkgPath := createNupkgArchive(t, nuspecContent)
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithResponse(200, `{}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath}, context)

	stdout := map[string]interface{}{}
	err := json.Unmarshal([]byte(result.StdOut), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize publish command result: %v", err)
	}
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status to be Succeeded, but got: %v", result.StdOut)
	}
	if stdout["error"] != nil {
		t.Errorf("Expected error to be nil, but got: %v", result.StdOut)
	}
	if stdout["name"] != "My Library" {
		t.Errorf("Expected name to be My Library, but got: %v", result.StdOut)
	}
	if stdout["version"] != "1.0.0" {
		t.Errorf("Expected version to be 1.0.0, but got: %v", result.StdOut)
	}
	if stdout["package"] == nil || stdout["package"] == "" {
		t.Errorf("Expected package not to be empty, but got: %v", result.StdOut)
	}
}

func TestPublishUploadsPackageToOrchestrator(t *testing.T) {
	nupkgPath := createNupkgArchive(t, nuspecContent)
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithResponse(200, `{}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath}, context)

	if result.RequestUrl != "/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage" {
		t.Errorf("Expected upload package request url, but got: %v", result.RequestUrl)
	}
	contentType := result.RequestHeader["content-type"]
	if !strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
		t.Errorf("Expected Content-Type header to be multipart/form-data, but got: %v", contentType)
	}
	expectedContentDisposition := fmt.Sprintf(`Content-Disposition: form-data; name="file"; filename="%s"`, filepath.Base(nupkgPath))
	if !strings.Contains(result.RequestBody, expectedContentDisposition) {
		t.Errorf("Expected request body to contain Content-Disposition, but got: %v", result.RequestBody)
	}
	if !strings.Contains(result.RequestBody, "Content-Type: application/octet-stream") {
		t.Errorf("Expected request body to contain Content-Type, but got: %v", result.RequestBody)
	}
	if !strings.Contains(result.RequestBody, "MyProcess.nuspec") {
		t.Errorf("Expected request body to contain nuspec file, but got: %v", result.RequestBody)
	}
}

func TestPublishUploadsLatestPackageFromDirectory(t *testing.T) {
	dir := createDirectory(t)
	archive1Path := filepath.Join(dir, "archive1.nupkg")
	archive2Path := filepath.Join(dir, "archive2.nupkg")
	writeNupkgArchive(t, archive1Path, nuspecContent)
	writeNupkgArchive(t, archive2Path, nuspecContent)

	err := os.Chtimes(archive1Path, time.Time{}, time.Now().Add(-5*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chtimes(archive2Path, time.Time{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithResponse(200, `{}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", dir}, context)

	stdout := map[string]interface{}{}
	err = json.Unmarshal([]byte(result.StdOut), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize publish command result: %v", err)
	}
	if !strings.HasSuffix(stdout["package"].(string), "archive2.nupkg") {
		t.Errorf("Expected publish to use latest nupkg package, but got: %v", result.StdOut)
	}
}

func TestPublishLargeFile(t *testing.T) {
	size := 10 * 1024 * 1024
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != size {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("Invalid size"))
			return
		}
		w.WriteHeader(201)
	}))
	defer srv.Close()

	nupkgPath := createLargeNupkgArchive(t, size)
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithResponse(200, `{"Uri":"`+srv.URL+`"}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
}

func createLargeNupkgArchive(t *testing.T, size int) string {
	path := createFile(t)
	archive, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)
	nuspecWriter, err := zipWriter.Create("MyProcess.nuspec")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.WriteString(nuspecWriter, nuspecContent)
	if err != nil {
		t.Fatal(err)
	}

	content, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   "Content.txt",
		Method: zip.Store,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = content.Write(make([]byte, size))
	if err != nil {
		t.Fatal(err)
	}
	err = zipWriter.Close()
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPublishWithDebugFlagOutputsRequestData(t *testing.T) {
	nupkgPath := createNupkgArchive(t, nuspecContent)
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithResponse(200, `{}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath, "--debug"}, context)

	if !strings.Contains(result.StdErr, "/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage") {
		t.Errorf("Expected stderr to show the upload package operation, but got: %v", result.StdErr)
	}
}

func TestPublishPackageAlreadyExistsReturnsFailed(t *testing.T) {
	nupkgPath := createNupkgArchive(t, `
<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
	<metadata minClientVersion="3.3">
	<id>MyProcess</id>
	<version>2.0.0</version>
	<title>My Process</title>
	</metadata>
</package>`)
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithResponse(409, `{}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath}, context)

	stdout := map[string]interface{}{}
	err := json.Unmarshal([]byte(result.StdOut), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize publish command result: %v", err)
	}
	if stdout["status"] != "Failed" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
	expectedError := fmt.Sprintf("Package '%s' already exists", filepath.Base(nupkgPath))
	if stdout["error"] != expectedError {
		t.Errorf("Expected error to be Package already exists, but got: %v", result.StdOut)
	}
	if stdout["name"] != "My Process" {
		t.Errorf("Expected name to be My Process, but got: %v", result.StdOut)
	}
	if stdout["version"] != "2.0.0" {
		t.Errorf("Expected version to be 2.0.0, but got: %v", result.StdOut)
	}
	if stdout["package"] == nil || stdout["package"] == "" {
		t.Errorf("Expected package not to be empty, but got: %v", result.StdOut)
	}
}

func TestPublishOrchestratorErrorReturnsError(t *testing.T) {
	nupkgPath := createNupkgArchive(t, nuspecContent)
	context := test.NewContextBuilder().
		WithDefinition("studio", studioDefinition).
		WithResponse(503, `{}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath}, context)

	if result.Error == nil || result.Error.Error() != "Service returned status code '503' and body '{}'" {
		t.Errorf("Expected orchestrator error, but got: %v", result.Error)
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
