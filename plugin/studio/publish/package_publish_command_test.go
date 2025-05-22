package publish

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
)

func TestPublishNoPackageFileReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", "not-found"}, context)

	if result.Error == nil || result.Error.Error() != "Package not found." {
		t.Errorf("Expected package not found error, but got: %v", result.Error)
	}
}

func TestPublishMissingOrganizationReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--tenant", "my-tenant", "--source", "my.nupkg"}, context)

	if result.Error == nil || result.Error.Error() != "Organization is not set" {
		t.Errorf("Expected organization is not set error, but got: %v", result.Error)
	}
}

func TestPublishMissingTenantReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--source", "my.nupkg"}, context)

	if result.Error == nil || result.Error.Error() != "Tenant is not set" {
		t.Errorf("Expected tenant is not set error, but got: %v", result.Error)
	}
}

func TestPublishInvalidPackageReturnsError(t *testing.T) {
	path := test.CreateTempFile(t, "invalid")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", path}, context)

	if result.Error == nil || !strings.HasPrefix(result.Error.Error(), "Could not read package") {
		t.Errorf("Expected package read error, but got: %v", result.Error)
	}
}

func TestPublishReturnsPackageMetadata(t *testing.T) {
	nupkgPath := createDefaultNupkgArchive(t)
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders?$filter=FullyQualifiedName eq 'Shared'", http.StatusOK, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/PackageFeeds/GetFolderFeed?folderId=938064", http.StatusOK, `null`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", http.StatusOK, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey eq 'MyProcess'", http.StatusOK, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases", http.StatusCreated, `{"name":"MyProcess","processKey":"MyProcess","processVersion":"1.0.195912597"}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status to be Succeeded, but got: %v", result.StdOut)
	}
	if stdout["error"] != nil {
		t.Errorf("Expected error to be nil, but got: %v", result.StdOut)
	}
	if stdout["name"] != "MyProcess" {
		t.Errorf("Expected name to be MyProcess, but got: %v", result.StdOut)
	}
	if stdout["description"] != "My Process" {
		t.Errorf("Expected description to be My Process, but got: %v", result.StdOut)
	}
	if stdout["version"] != "1.0.0" {
		t.Errorf("Expected version to be 1.0.0, but got: %v", result.StdOut)
	}
	if stdout["package"] == nil || stdout["package"] == "" {
		t.Errorf("Expected package not to be empty, but got: %v", result.StdOut)
	}
}

func TestPublishUploadsPackageToOrchestrator(t *testing.T) {
	nupkgPath := createDefaultNupkgArchive(t)
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders?$filter=FullyQualifiedName eq 'Shared'", http.StatusOK, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/PackageFeeds/GetFolderFeed?folderId=938064", http.StatusOK, `null`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", http.StatusInternalServerError, `{}`).
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
	dir := t.TempDir()
	archive1Path := filepath.Join(dir, "archive1.nupkg")
	archive2Path := filepath.Join(dir, "archive2.nupkg")

	err := studio.NewNupkgWriter(archive1Path).
		WithNuspec(*studio.NewNuspec("MyProcess", "My Process", "1.0.0")).
		Write()
	if err != nil {
		t.Fatal(err)
	}
	err = studio.NewNupkgWriter(archive2Path).
		WithNuspec(*studio.NewNuspec("MyProcess", "My Process", "1.0.0")).
		Write()
	if err != nil {
		t.Fatal(err)
	}

	err = os.Chtimes(archive1Path, time.Time{}, time.Now().Add(time.Duration(-5)*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chtimes(archive2Path, time.Time{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders?$filter=FullyQualifiedName eq 'Shared'", http.StatusOK, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/PackageFeeds/GetFolderFeed?folderId=938064", http.StatusOK, `null`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", http.StatusOK, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey eq 'MyProcess'", http.StatusOK, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases", http.StatusCreated, `{"name":"MyProcess","processKey":"MyProcess","processVersion":"1.0.195912597"}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", dir}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	if !strings.HasSuffix(stdout["package"].(string), "archive2.nupkg") {
		t.Errorf("Expected publish to use latest nupkg package, but got: %v", result.StdOut)
	}
}

func TestPublishLargeFile(t *testing.T) {
	size := 10 * 1024 * 1024
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) != size {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Invalid size"))
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	nupkgPath := createLargeNupkgArchive(t, size)
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders?$filter=FullyQualifiedName eq 'Shared'", http.StatusOK, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/PackageFeeds/GetFolderFeed?folderId=938064", http.StatusOK, `null`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", http.StatusOK, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey eq 'MyProcess'", http.StatusOK, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases", http.StatusCreated, `{"name":"MyProcess","processKey":"MyProcess","processVersion":"1.0.195912597"}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
}

func TestPublishWithDebugFlagOutputsRequestData(t *testing.T) {
	nupkgPath := createDefaultNupkgArchive(t)
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders?$filter=FullyQualifiedName eq 'Shared'", http.StatusOK, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/PackageFeeds/GetFolderFeed?folderId=938064", http.StatusOK, `null`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", http.StatusOK, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey eq 'MyProcess'", http.StatusOK, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases", http.StatusCreated, `{"name":"MyProcess","processKey":"MyProcess","processVersion":"1.0.195912597"}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath, "--debug"}, context)

	if !strings.Contains(result.StdErr, "/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage") {
		t.Errorf("Expected stderr to show the upload package operation, but got: %v", result.StdErr)
	}
}

func TestPublishPackageAlreadyExistsReturnsFailed(t *testing.T) {
	nupkgPath := createNupkgArchive(t, *studio.NewNuspec("MyProcess", "My Process", "2.0.0"))
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders?$filter=FullyQualifiedName eq 'Shared'", http.StatusOK, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/PackageFeeds/GetFolderFeed?folderId=938064", http.StatusOK, `null`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", http.StatusConflict, `{}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Failed" {
		t.Errorf("Expected status to be Failed, but got: %v", result.StdOut)
	}
	expectedError := fmt.Sprintf("Package '%s' already exists", filepath.Base(nupkgPath))
	if stdout["error"] != expectedError {
		t.Errorf("Expected error to be Package already exists, but got: %v", result.StdOut)
	}
	if stdout["name"] != "MyProcess" {
		t.Errorf("Expected name to be MyProcess, but got: %v", result.StdOut)
	}
	if stdout["description"] != "My Process" {
		t.Errorf("Expected description to be My Process, but got: %v", result.StdOut)
	}
	if stdout["version"] != "2.0.0" {
		t.Errorf("Expected version to be 2.0.0, but got: %v", result.StdOut)
	}
	if stdout["package"] == nil || stdout["package"] == "" {
		t.Errorf("Expected package not to be empty, but got: %v", result.StdOut)
	}
}

func TestPublishOrchestratorErrorReturnsError(t *testing.T) {
	nupkgPath := createDefaultNupkgArchive(t)
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithResponse(http.StatusServiceUnavailable, `{}`).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath}, context)

	if result.Error == nil || result.Error.Error() != "Service returned status code '503' and body '{}'" {
		t.Errorf("Expected orchestrator error, but got: %v", result.Error)
	}
}

func TestPublishUsesProvidedFolderId(t *testing.T) {
	nupkgPath := createDefaultNupkgArchive(t)
	header := map[string]string{}
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders?$filter=FullyQualifiedName eq '12345' or Id eq 12345", http.StatusOK, `{"value":[{"Id":12345,"FullyQualifiedName":"MyFolder"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/PackageFeeds/GetFolderFeed?folderId=12345", http.StatusOK, `null`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", http.StatusOK, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey eq 'MyProcess'", http.StatusOK, `{"value":[]}`).
		WithResponseHandler(func(request test.RequestData) test.ResponseData {
			header = request.Header
			return test.ResponseData{Status: http.StatusOK, Body: ""}
		}).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath, "--folder-id", "12345"}, context)

	folderId := header["x-uipath-organizationunitid"]
	if folderId != "12345" {
		t.Errorf("Expected x-uipath-organizationunitid header from argument, but got: '%s'", folderId)
	}
}

func TestPublishUsesProvidedFolder(t *testing.T) {
	nupkgPath := createDefaultNupkgArchive(t)
	header := map[string]string{}
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders?$filter=FullyQualifiedName eq 'MyFolder'", http.StatusOK, `{"value":[{"Id":12345,"FullyQualifiedName":"MyFolder"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/PackageFeeds/GetFolderFeed?folderId=12345", http.StatusOK, `null`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", http.StatusOK, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey eq 'MyProcess'", http.StatusOK, `{"value":[]}`).
		WithResponseHandler(func(request test.RequestData) test.ResponseData {
			header = request.Header
			return test.ResponseData{Status: http.StatusOK, Body: ""}
		}).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath, "--folder", "MyFolder"}, context)

	folderId := header["x-uipath-organizationunitid"]
	if folderId != "12345" {
		t.Errorf("Expected x-uipath-organizationunitid header from argument, but got: '%s'", folderId)
	}
}

func TestPublishUsesFolderFeedWhenAvailable(t *testing.T) {
	nupkgPath := createDefaultNupkgArchive(t)
	header := map[string]string{}
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders?$filter=FullyQualifiedName eq '12345' or Id eq 12345", http.StatusOK, `{"value":[{"Id":12345,"FullyQualifiedName":"MyFolder"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/PackageFeeds/GetFolderFeed?folderId=12345", http.StatusOK, `8e00fda5-6124-43ca-b8c8-5d812589e567`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage?feedId=8e00fda5-6124-43ca-b8c8-5d812589e567", http.StatusOK, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey eq 'MyProcess'", http.StatusOK, `{"value":[]}`).
		WithResponseHandler(func(request test.RequestData) test.ResponseData {
			header = request.Header
			return test.ResponseData{Status: http.StatusOK, Body: ""}
		}).
		WithCommandPlugin(NewPackagePublishCommand()).
		Build()

	test.RunCli([]string{"studio", "package", "publish", "--organization", "my-org", "--tenant", "my-tenant", "--source", nupkgPath, "--folder-id", "12345"}, context)

	folderId := header["x-uipath-organizationunitid"]
	if folderId != "12345" {
		t.Errorf("Expected x-uipath-organizationunitid header from argument, but got: '%s'", folderId)
	}
}

func createDefaultNupkgArchive(t *testing.T) string {
	return createNupkgArchive(t, *studio.NewNuspec("MyProcess", "My Process", "1.0.0"))
}

func createNupkgArchive(t *testing.T, nuspec studio.Nuspec) string {
	path := test.TempFile(t)
	err := studio.NewNupkgWriter(path).
		WithNuspec(nuspec).
		Write()
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func createLargeNupkgArchive(t *testing.T, size int) string {
	path := test.TempFile(t)
	err := studio.NewNupkgWriter(path).
		WithNuspec(*studio.NewNuspec("MyProcess", "My Process", "1.0.0")).
		WithFile("Content.txt", make([]byte, size)).
		Write()
	if err != nil {
		t.Fatal(err)
	}
	return path
}
