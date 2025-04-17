package download

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/test"
)

func TestDownloadWithoutFolderIdParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewDownloadCommand()).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Argument --folder-id is missing") {
		t.Errorf("Expected stderr to show that folder-id parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutKeyParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewDownloadCommand()).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Argument --key is missing") {
		t.Errorf("Expected stderr to show that key parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutPathParameterShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewDownloadCommand()).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2"}, context)

	if !strings.Contains(result.StdErr, "Argument --path is missing") {
		t.Errorf("Expected stderr to show that path parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutOrganizationShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewDownloadCommand()).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Organization is not set") {
		t.Errorf("Expected stderr to show that organization parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutTenantShowsValidationError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(NewDownloadCommand()).
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
		WithCommandPlugin(NewDownloadCommand()).
		WithResponse(http.StatusBadRequest, "validation error").
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Orchestrator returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that orchestrator call failed, but got: %v", result.StdErr)
	}
}

func TestDownloadSuccessfully(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Wrong http method"))
			return
		}
		w.WriteHeader(http.StatusOK)
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
		WithCommandPlugin(NewDownloadCommand()).
		WithResponse(http.StatusOK, `{"Uri":"`+srv.URL+`"}`).
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
		w.Header().Set("Content-Length", strconv.Itoa(size))
		w.WriteHeader(http.StatusOK)
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
		WithCommandPlugin(NewDownloadCommand()).
		WithResponse(http.StatusOK, `{"Uri":"`+srv.URL+`"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--organization", "myorg", "--tenant", "mytenant", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if len(result.StdOut) < 10*1024*1024 {
		t.Errorf("Expected stdout to show file content, but got: %v", len(result.StdOut))
	}
}

func TestDownloadWithDebugOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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
		WithCommandPlugin(NewDownloadCommand()).
		WithResponse(http.StatusOK, `{"Uri":"`+srv.URL+`/download/file.txt"}`).
		Build()

	result := test.RunCli([]string{"orchestrator", "buckets", "download", "--debug", "--organization", "myorg", "--tenant", "mytenant", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "/myorg/mytenant/orchestrator_/odata/Buckets(2)/UiPath.Server.Configuration.OData.GetReadUri?path=file.txt") {
		t.Errorf("Expected stderr to contain first request to get read uri, but got: %v", result.StdErr)
	}
	if !strings.Contains(result.StdErr, "GET "+srv.URL+"/download/file.txt") {
		t.Errorf("Expected stderr to contain download request, but got: %v", result.StdErr)
	}
}
