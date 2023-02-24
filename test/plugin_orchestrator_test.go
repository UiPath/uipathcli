package test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	plugin_orchestrator "github.com/UiPath/uipathcli/plugin/orchestrator"
)

func TestUploadWithoutFolderIdParameterShowsValidationError(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(plugin_orchestrator.UploadCommand{}).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "upload", "--key", "2", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Argument --folder-id is missing") {
		t.Errorf("Expected stderr to show that folder-id parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutKeyParameterShowsValidationError(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(plugin_orchestrator.UploadCommand{}).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Argument --key is missing") {
		t.Errorf("Expected stderr to show that key parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutPathParameterShowsValidationError(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(plugin_orchestrator.UploadCommand{}).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Argument --path is missing") {
		t.Errorf("Expected stderr to show that path parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutFileParameterShowsValidationError(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(plugin_orchestrator.UploadCommand{}).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Argument --file is missing") {
		t.Errorf("Expected stderr to show that file parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadFileDoesNotExistShowsValidationError(t *testing.T) {
	config := `profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
`

	context := NewContextBuilder().
		WithConfig(config).
		WithDefinition("orchestrator", "").
		WithCommandPlugin(plugin_orchestrator.UploadCommand{}).
		WithResponse(200, `{"Uri":"http://localhost"}`).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Error sending request: File 'does-not-exist' not found") {
		t.Errorf("Expected stderr to show that file was not found, but got: %v", result.StdErr)
	}
}

func TestUploadWithoutOrganizationShowsValidationError(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(plugin_orchestrator.UploadCommand{}).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", "hello-world"}, context)

	if !strings.Contains(result.StdErr, "Could not find 'organization' parameter") {
		t.Errorf("Expected stderr to show that organization parameter is missing, but got: %v", result.StdErr)
	}
}

func TestUploadWithFailedResponseReturnsError(t *testing.T) {
	path := createFile(t)
	os.WriteFile(path, []byte("hello-world"), 0644)

	config := `profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
`

	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithConfig(config).
		WithCommandPlugin(plugin_orchestrator.UploadCommand{}).
		WithResponse(400, "validation error").
		Build()

	result := runCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", path}, context)

	if !strings.Contains(result.StdErr, "Orchestrator returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that orchestrator call failed, but got: %v", result.StdErr)
	}
}

func TestUploadSuccessfully(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			w.WriteHeader(500)
			w.Write([]byte("Wrong http method"))
			return
		}
		if r.Header["X-Ms-Blob-Type"][0] != "BlockBlob" {
			w.WriteHeader(500)
			w.Write([]byte("Missing header x-ms-blob-type"))
			return
		}
		body, _ := io.ReadAll(r.Body)
		requestBody := string(body)
		if requestBody != "hello-world" {
			w.WriteHeader(500)
			w.Write([]byte("File content not found"))
			return
		}
		w.WriteHeader(201)
	}))
	defer srv.Close()

	path := createFile(t)
	os.WriteFile(path, []byte("hello-world"), 0644)

	config := `profiles:
- name: default
  path:
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

	context := NewContextBuilder().
		WithDefinition("orchestrator", definition).
		WithConfig(config).
		WithCommandPlugin(plugin_orchestrator.UploadCommand{}).
		WithResponse(200, `{"Uri":"`+srv.URL+`"}`).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "upload", "--folder-id", "1", "--key", "2", "--path", "file.txt", "--file", path}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	if result.StdErr != "" {
		t.Errorf("Expected stderr to be empty, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutFolderIdParameterShowsValidationError(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(plugin_orchestrator.DownloadCommand{}).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "download", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Argument --folder-id is missing") {
		t.Errorf("Expected stderr to show that folder-id parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutKeyParameterShowsValidationError(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(plugin_orchestrator.DownloadCommand{}).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Argument --key is missing") {
		t.Errorf("Expected stderr to show that key parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutPathParameterShowsValidationError(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(plugin_orchestrator.DownloadCommand{}).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2"}, context)

	if !strings.Contains(result.StdErr, "Argument --path is missing") {
		t.Errorf("Expected stderr to show that path parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithoutOrganizationShowsValidationError(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithCommandPlugin(plugin_orchestrator.DownloadCommand{}).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Could not find 'organization' parameter") {
		t.Errorf("Expected stderr to show that organization parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDownloadWithFailedResponseReturnsError(t *testing.T) {
	config := `profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
`

	context := NewContextBuilder().
		WithDefinition("orchestrator", "").
		WithConfig(config).
		WithCommandPlugin(plugin_orchestrator.DownloadCommand{}).
		WithResponse(400, "validation error").
		Build()

	result := runCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

	if !strings.Contains(result.StdErr, "Orchestrator returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that orchestrator call failed, but got: %v", result.StdErr)
	}
}

func TestDownloadSuccessfully(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(500)
			w.Write([]byte("Wrong http method"))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("hello-world"))
	}))
	defer srv.Close()

	config := `profiles:
- name: default
  path:
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

	context := NewContextBuilder().
		WithDefinition("orchestrator", definition).
		WithConfig(config).
		WithCommandPlugin(plugin_orchestrator.DownloadCommand{}).
		WithResponse(200, `{"Uri":"`+srv.URL+`"}`).
		Build()

	result := runCli([]string{"orchestrator", "buckets", "download", "--folder-id", "1", "--key", "2", "--path", "file.txt"}, context)

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
