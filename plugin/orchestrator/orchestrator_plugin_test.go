package orchestrator

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/test"
)

func TestUploadWithoutFolderIdParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(UploadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--key", "2", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Argument --folder-id is missing") {
		t.Errorf("Expected stderr to show that folder-id parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutKeyParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(UploadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Argument --key is missing") {
		t.Errorf("Expected stderr to show that key parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithInvalidFolderIdParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(UploadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "invalid", "--key", "2", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Cannot convert 'folder-id' value 'invalid' to integer") {
		t.Errorf("Expected stderr to show that folder id cannot be converted to integer, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutPathParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(UploadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Argument --path is missing") {
		t.Errorf("Expected stderr to show that path parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutFileParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(UploadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Argument --file is missing") {
		t.Errorf("Expected stderr to show that file parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadFileDoesNotExistShowsValidationError(t *testing.T) {
	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	context := test.NewContextBuilder().
		WithConfig(config).
		WithDefinition("orchestrator", "").
		WithCommandPlugin(UploadCommand{}).
		WithResponse(200, `{"Uri":"http://localhost"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Error sending request: File 'does-not-exist' not found") {
		t.Errorf("Expected stderr to show that file was not found, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutOrganizationShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(UploadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", "hello-world"}, context)

	if !strings.Contains(result.StdErr, "Organization is not set") {
		t.Errorf("Expected stderr to show that organization parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutTenantShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(UploadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--organization", "myorg", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", "hello-world"}, context)

	if !strings.Contains(result.StdErr, "Tenant is not set") {
		t.Errorf("Expected stderr to show that tenant parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithFailedResponseReturnsError(t *testing.T) {
	path := createFile(t)
	writeFile(path, []byte("hello-world"))

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithConfig(config).
		WithCommandPlugin(UploadCommand{}).
		WithResponse(400, "validation error").
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", path}, context)

	if !strings.Contains(result.StdErr, "Orchestrator returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that orchestrator call failed, but got: %v", result.StdErr)
	}
}

func TestUploadSuccessfully(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("Wrong http method"))
			return
		}
		if r.Header["X-Ms-Blob-Type"][0] != "BlockBlob" {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("Missing header x-ms-blob-type"))
			return
		}
		body, _ := io.ReadAll(r.Body)
		requestBody := string(body)
		if requestBody != "hello-world" {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("File content not found"))
			return
		}
		w.WriteHeader(201)
	}))
	defer srv.Close()

	path := createFile(t)
	writeFile(path, []byte("hello-world"))

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/orchestrator_
  description: The production url
  variables:
    organization:
      description: The organization name (or id)
      default: my-org
    tenant:
      description: The tenant name (or id)
      default: my-tenant
`

	context := test.NewContextBuilder().
		WithDefinition("orchestrator", definition).
		WithConfig(config).
		WithCommandPlugin(UploadCommand{}).
		WithResponse(200, `{"Uri":"`+srv.URL+`"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", path}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	if result.StdErr != "" {
		t.Errorf("Expected stderr to be empty, but got: %v", result.StdErr)
	}
}

func TestUploadLargeFile(t *testing.T) {
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

	path := createFile(t)
	writeFile(path, make([]byte, size))

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/orchestrator_
  description: The production url
  variables:
    organization:
      description: The organization name (or id)
      default: my-org
    tenant:
      description: The tenant name (or id)
      default: my-tenant
`

	context := test.NewContextBuilder().
		WithDefinition("orchestrator", definition).
		WithCommandPlugin(UploadCommand{}).
		WithResponse(200, `{"Uri":"`+srv.URL+`"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--organization", "myorg", "--tenant", "mytenant", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", path}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
}

func TestUploadWithDebugOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
	}))
	defer srv.Close()

	path := createFile(t)
	writeFile(path, []byte("hello-world"))

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/orchestrator_
  description: The production url
  variables:
    organization:
      description: The organization name (or id)
      default: my-org
    tenant:
      description: The tenant name (or id)
      default: my-tenant
`

	context := test.NewContextBuilder().
		WithDefinition("orchestrator", definition).
		WithCommandPlugin(UploadCommand{}).
		WithResponse(200, `{"Uri":"`+srv.URL+"/upload/file.txt"+`"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--debug", "--organization", "myorg", "--tenant", "mytenant", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", path}, context)

	if !strings.Contains(result.StdErr, "/myorg/mytenant/orchestrator_/odata/Buckets(2)/UiPath.Server.Configuration.OData.GetWriteUri?path=file.txt") {
		t.Errorf("Expected stderr to contain first request to get write uri, but got: %v", result.StdErr)
	}
	if !strings.Contains(result.StdErr, "PUT "+srv.URL+"/upload/file.txt") {
		t.Errorf("Expected stderr to contain upload request, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutFolderIdParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(DownloadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Argument --folder-id is missing") {
		t.Errorf("Expected stderr to show that folder-id parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutKeyParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(DownloadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Argument --key is missing") {
		t.Errorf("Expected stderr to show that key parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutPathParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(DownloadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2"}, context)

	if !strings.Contains(result.StdErr, "Argument --path is missing") {
		t.Errorf("Expected stderr to show that path parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutOrganizationShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(DownloadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Organization is not set") {
		t.Errorf("Expected stderr to show that organization parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutTenantShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(DownloadCommand{}).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--organization", "myorg", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Tenant is not set") {
		t.Errorf("Expected stderr to show that tenant parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithFailedResponseReturnsError(t *testing.T) {
	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithConfig(config).
		WithCommandPlugin(DownloadCommand{}).
		WithResponse(400, "validation error").
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Orchestrator returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that orchestrator call failed, but got: %v", result.StdErr)
	}
}

func TestDownloadSuccessfully(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("Wrong http method"))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("hello-world"))
	}))
	defer srv.Close()

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/orchestrator_
  description: The production url
  variables:
    organization:
      description: The organization name (or id)
      default: my-org
    tenant:
      description: The tenant name (or id)
      default: my-tenant
`

	context := test.NewContextBuilder().
		WithDefinition("orchestrator", definition).
		WithConfig(config).
		WithCommandPlugin(DownloadCommand{}).
		WithResponse(200, `{"Uri":"`+srv.URL+`"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	if result.StdErr != "" {
		t.Errorf("Expected stderr to be empty, but got: %v", result.StdErr)
	}
	if result.StdOut != "hello-world" {
		t.Errorf("Expected stdout to show file content, but got: %v", result.StdOut)
	}
}

func TestDownloadLargeFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		size := 10 * 1024 * 1024
		w.Header().Add("content-length", strconv.Itoa(size))
		w.WriteHeader(200)
		_, _ = w.Write(make([]byte, size))
	}))
	defer srv.Close()

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/orchestrator_
  description: The production url
  variables:
    organization:
      description: The organization name (or id)
      default: my-org
    tenant:
      description: The tenant name (or id)
      default: my-tenant
`

	context := test.NewContextBuilder().
		WithDefinition("orchestrator", definition).
		WithCommandPlugin(DownloadCommand{}).
		WithResponse(200, `{"Uri":"`+srv.URL+`"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--organization", "myorg", "--tenant", "mytenant", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if len(result.StdOut) < 10*1024*1024 {
		t.Errorf("Expected stdout to show file content, but got: %v", len(result.StdOut))
	}
}

func TestDownloadWithDebugOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("hello-world"))
	}))
	defer srv.Close()

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/orchestrator_
  description: The production url
  variables:
    organization:
      description: The organization name (or id)
      default: my-org
    tenant:
      description: The tenant name (or id)
      default: my-tenant
`

	context := test.NewContextBuilder().
		WithDefinition("orchestrator", definition).
		WithCommandPlugin(DownloadCommand{}).
		WithResponse(200, `{"Uri":"`+srv.URL+`/download/file.txt"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--debug", "--organization", "myorg", "--tenant", "mytenant", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "/myorg/mytenant/orchestrator_/odata/Buckets(2)/UiPath.Server.Configuration.OData.GetReadUri?path=file.txt") {
		t.Errorf("Expected stderr to contain first request to get read uri, but got: %v", result.StdErr)
	}
	if !strings.Contains(result.StdErr, "GET "+srv.URL+"/download/file.txt") {
		t.Errorf("Expected stderr to contain download request, but got: %v", result.StdErr)
	}
}

func createFile(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "uipath-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tempFile.Name()) })
	return tempFile.Name()
}

func writeFile(name string, data []byte) {
	err := os.WriteFile(name, data, 0600)
	if err != nil {
		panic(fmt.Errorf("Error writing file '%s': %w", name, err))
	}
}
