package upload

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/test"
)

func TestUploadWithoutFolderIdParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewUploadCommand()).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--key", "2", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Argument --folder-id is missing") {
		t.Errorf("Expected stderr to show that folder-id parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutKeyParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewUploadCommand()).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Argument --key is missing") {
		t.Errorf("Expected stderr to show that key parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithInvalidFolderIdParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewUploadCommand()).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "invalid", "--key", "2", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Cannot convert 'folder-id' value 'invalid' to integer") {
		t.Errorf("Expected stderr to show that folder id cannot be converted to integer, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutPathParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewUploadCommand()).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Argument --path is missing") {
		t.Errorf("Expected stderr to show that path parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutFileParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewUploadCommand()).
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
		WithCommandPlugin(NewUploadCommand()).
		WithResponse(http.StatusOK, `{"Uri":"http://localhost"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Error sending request: File 'does-not-exist' not found") {
		t.Errorf("Expected stderr to show that file was not found, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutOrganizationShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewUploadCommand()).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", "hello-world"}, context)

	if !strings.Contains(result.StdErr, "Organization is not set") {
		t.Errorf("Expected stderr to show that organization parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutTenantShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewUploadCommand()).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--organization", "myorg", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", "hello-world"}, context)

	if !strings.Contains(result.StdErr, "Tenant is not set") {
		t.Errorf("Expected stderr to show that tenant parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithFailedResponseReturnsError(t *testing.T) {
	path := test.CreateTempFile(t, "hello-world")

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithConfig(config).
		WithCommandPlugin(NewUploadCommand()).
		WithResponse(http.StatusBadRequest, "validation error").
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", path}, context)

	if !strings.Contains(result.StdErr, "Orchestrator returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that orchestrator call failed, but got: %v", result.StdErr)
	}
}

func TestUploadSuccessfully(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Wrong http method"))
			return
		}
		if r.Header["X-Ms-Blob-Type"][0] != "BlockBlob" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Missing header x-ms-blob-type"))
			return
		}
		body, _ := io.ReadAll(r.Body)
		requestBody := string(body)
		if requestBody != "hello-world" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("File content not found"))
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	path := test.CreateTempFile(t, "hello-world")

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
		WithCommandPlugin(NewUploadCommand()).
		WithResponse(http.StatusOK, `{"Uri":"`+srv.URL+`"}`).
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
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Invalid size"))
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	path := test.CreateTempFileBinary(t, make([]byte, size))

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
		WithCommandPlugin(NewUploadCommand()).
		WithResponse(http.StatusOK, `{"Uri":"`+srv.URL+`"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--organization", "myorg", "--tenant", "mytenant", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", path}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
}

func TestUploadWithDebugOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	path := test.CreateTempFile(t, "hello-world")

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
		WithCommandPlugin(NewUploadCommand()).
		WithResponse(http.StatusOK, `{"Uri":"`+srv.URL+"/upload/file.txt"+`"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "upload", "--debug", "--organization", "myorg", "--tenant", "mytenant", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", path}, context)

	if !strings.Contains(result.StdErr, "/myorg/mytenant/orchestrator_/odata/Buckets(2)/UiPath.Server.Configuration.OData.GetWriteUri?path=file.txt") {
		t.Errorf("Expected stderr to contain first request to get write uri, but got: %v", result.StdErr)
	}
	if !strings.Contains(result.StdErr, "PUT "+srv.URL+"/upload/file.txt") {
		t.Errorf("Expected stderr to contain upload request, but got: %v", result.StdErr)
	}
}
