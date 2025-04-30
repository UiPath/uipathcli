package digitzer

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/test"
)

func TestDigitizeWithoutFileParameterShowsValidationError(t *testing.T) {
	definition := `
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithCommandPlugin(NewDigitizeCommand()).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234"}, context)

	if !strings.Contains(result.StdErr, "Argument --file is missing") {
		t.Errorf("Expected stderr to show that file parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDigitizeFileDoesNotExistShowsValidationError(t *testing.T) {
	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithConfig(config).
		WithDefinition("du", definition).
		WithCommandPlugin(NewDigitizeCommand()).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--file", "does-not-exist"}, context)

	if !strings.Contains(result.StdErr, "Error sending request: File 'does-not-exist' not found") {
		t.Errorf("Expected stderr to show that file was not found, but got: %v", result.StdErr)
	}
}

func TestDigitizeWithoutOrganizationShowsValidationError(t *testing.T) {
	definition := `
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithCommandPlugin(NewDigitizeCommand()).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--file", "hello-world"}, context)

	if !strings.Contains(result.StdErr, "Organization is not set") {
		t.Errorf("Expected stderr to show that organization parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDigitizeWithoutTenantShowsValidationError(t *testing.T) {
	definition := `
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithCommandPlugin(NewDigitizeCommand()).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--organization", "myorg", "--project-id", "1234", "--file", "hello-world"}, context)

	if !strings.Contains(result.StdErr, "Tenant is not set") {
		t.Errorf("Expected stderr to show that tenant parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDigitizeWithFailedResponseReturnsError(t *testing.T) {
	path := test.CreateTempFile(t, "hello-world")

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithConfig(config).
		WithCommandPlugin(NewDigitizeCommand()).
		WithResponse(http.StatusBadRequest, "validation error").
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--file", path}, context)

	if !strings.Contains(result.StdErr, "Digitizer returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that digitizer call failed, but got: %v", result.StdErr)
	}
}

func TestDigitizeWithFailedResultResponseReturnsError(t *testing.T) {
	path := test.CreateTempFile(t, "hello-world")

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithConfig(config).
		WithCommandPlugin(NewDigitizeCommand()).
		WithResponse(http.StatusAccepted, `{"documentId":"04908673-2b65-4647-8ab3-dde8a3aa7885"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/04908673-2b65-4647-8ab3-dde8a3aa7885?api-version=1", http.StatusBadRequest, `validation error`).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--file", path}, context)

	if !strings.Contains(result.StdErr, "Digitizer returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that digitizer call failed, but got: %v", result.StdErr)
	}
}

func TestDigitizeWithoutProjectIdUsesDefaultProject(t *testing.T) {
	path := test.CreateTempFile(t, "hello-world")

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/du_/api/framework
  description: The production url
  variables:
    organization:
      description: The organization name (or id)
      default: my-org
    tenant:
      description: The tenant name (or id)
      default: my-tenant
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithConfig(config).
		WithCommandPlugin(NewDigitizeCommand()).
		WithResponse(http.StatusAccepted, `{"documentId":"648ea1c2-7dbe-42a8-b112-6474d07e61c1"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/00000000-0000-0000-0000-000000000000/digitization/result/648ea1c2-7dbe-42a8-b112-6474d07e61c1?api-version=1", http.StatusOK, `{"status":"Done"}`).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--file", path}, context)

	expectedResult := `{
  "status": "Done"
}
`
	if result.StdOut != expectedResult {
		t.Errorf("Expected stdout to show the digitize result, but got: %v", result.StdOut)
	}
}

func TestDigitizeSuccessfully(t *testing.T) {
	path := test.CreateTempFile(t, "hello-world")

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/du_/api/framework
  description: The production url
  variables:
    organization:
      description: The organization name (or id)
      default: my-org
    tenant:
      description: The tenant name (or id)
      default: my-tenant
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithConfig(config).
		WithCommandPlugin(NewDigitizeCommand()).
		WithResponse(http.StatusAccepted, `{"documentId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", http.StatusOK, `{"status":"Done"}`).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--file", path}, context)

	expectedResult := `{
  "status": "Done"
}
`
	if result.StdOut != expectedResult {
		t.Errorf("Expected stdout to show the digitize result, but got: %v", result.StdOut)
	}
}

func TestDigitizeSuccessfullyWithDebugFlag(t *testing.T) {
	path := test.CreateTempFile(t, "hello-world")

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithConfig(config).
		WithCommandPlugin(NewDigitizeCommand()).
		WithResponse(http.StatusAccepted, `{"documentId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", http.StatusOK, `{"pages":[],"status":"Done"}`).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--file", path, "--debug"}, context)

	expected := `{
  "pages": [],
  "status": "Done"
}
`
	if result.StdOut != expected {
		t.Errorf("Expected stdout to show the digitize result, but got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdErr, "/digitization/start") {
		t.Errorf("Expected stderr to show the start digitization operation, but got: %v", result.StdErr)
	}
	if !strings.Contains(result.StdErr, "/digitization/result/eb80e441-05de-4a13-9aaa-f65b1babba05") {
		t.Errorf("Expected stderr to show the get digitization result operation, but got: %v", result.StdErr)
	}
}

func TestDigitizeSuccessfullyWithStdIn(t *testing.T) {
	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
paths:
  /digitize:
    get:
      operationId: digitize
`
	stdIn := bytes.Buffer{}
	stdIn.WriteString("hello-world")
	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithConfig(config).
		WithCommandPlugin(NewDigitizeCommand()).
		WithStdIn(stdIn).
		WithResponse(http.StatusAccepted, `{"documentId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", http.StatusOK, `{"status":"Done"}`).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--content-type", "application/pdf", "--file", "-"}, context)

	expectedResult := `{
  "status": "Done"
}
`
	if result.StdOut != expectedResult {
		t.Errorf("Expected stdout to show the digitize result, but got: %v", result.StdOut)
	}
}

func TestDigitizeLargeFileSuccessfully(t *testing.T) {
	path := test.CreateTempFileBinary(t, make([]byte, 10*1024*1024))

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/du_/api/framework
  description: The production url
  variables:
    organization:
      description: The organization name (or id)
      default: my-org
    tenant:
      description: The tenant name (or id)
      default: my-tenant
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithConfig(config).
		WithCommandPlugin(NewDigitizeCommand()).
		WithResponse(http.StatusAccepted, `{"documentId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", http.StatusOK, `{"status":"Done"}`).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--file", path}, context)

	expectedResult := `{
  "status": "Done"
}
`
	if result.StdOut != expectedResult {
		t.Errorf("Expected stdout to show the digitize result, but got: %v", result.StdOut)
	}
}

func TestDigitizeSuccessfullyWithCustomHeader(t *testing.T) {
	path := test.CreateTempFile(t, "hello-world")

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  header:
    x-custom-header: my-custom-value
`

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/du_/api/framework
  description: The production url
  variables:
    organization:
      description: The organization name (or id)
      default: my-org
    tenant:
      description: The tenant name (or id)
      default: my-tenant
paths:
  /digitize:
    get:
      operationId: digitize
`

	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithConfig(config).
		WithCommandPlugin(NewDigitizeCommand()).
		WithResponse(http.StatusAccepted, `{"documentId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", http.StatusOK, `{"status":"Done"}`).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--file", path}, context)

	if result.RequestHeader["x-custom-header"] != "my-custom-value" {
		t.Errorf("Expected HTTP calls to contain custom config header, but got: %v", result.RequestHeader)
	}
}
