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
		WithResponse(http.StatusOK, `{}`).
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
	if stdout["name"] != "My Process" {
		t.Errorf("Expected name to be My Process, but got: %v", result.StdOut)
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
		WithResponse(http.StatusOK, `{}`).
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

	err = os.Chtimes(archive1Path, time.Time{}, time.Now().Add(-5*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	err = os.Chtimes(archive2Path, time.Time{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithResponse(http.StatusOK, `{}`).
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
		WithResponse(http.StatusOK, `{"Uri":"`+srv.URL+`"}`).
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
		WithResponse(http.StatusOK, `{}`).
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
		WithResponse(http.StatusConflict, `{}`).
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
