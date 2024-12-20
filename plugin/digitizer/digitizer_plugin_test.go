package digitzer

import (
	"bytes"
	"fmt"
	"os"
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
		WithCommandPlugin(DigitizeCommand{}).
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
		WithCommandPlugin(DigitizeCommand{}).
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
		WithCommandPlugin(DigitizeCommand{}).
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
		WithCommandPlugin(DigitizeCommand{}).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--organization", "myorg", "--project-id", "1234", "--file", "hello-world"}, context)

	if !strings.Contains(result.StdErr, "Tenant is not set") {
		t.Errorf("Expected stderr to show that tenant parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDigitizeWithFailedResponseReturnsError(t *testing.T) {
	path := createFile(t)
	writeFile(path, []byte("hello-world"))

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
		WithCommandPlugin(DigitizeCommand{}).
		WithResponse(400, "validation error").
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--file", path}, context)

	if !strings.Contains(result.StdErr, "Digitizer returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that digitizer call failed, but got: %v", result.StdErr)
	}
}

func TestDigitizeWithFailedResultResponseReturnsError(t *testing.T) {
	path := createFile(t)
	writeFile(path, []byte("hello-world"))

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
		WithCommandPlugin(DigitizeCommand{}).
		WithResponse(202, `{"documentId":"04908673-2b65-4647-8ab3-dde8a3aa7885"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/04908673-2b65-4647-8ab3-dde8a3aa7885?api-version=1", 400, `validation error`).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--project-id", "1234", "--file", path}, context)

	if !strings.Contains(result.StdErr, "Digitizer returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that digitizer call failed, but got: %v", result.StdErr)
	}
}

func TestDigitizeWithoutProjectIdUsesDefaultProject(t *testing.T) {
	path := createFile(t)
	writeFile(path, []byte("hello-world"))

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
		WithCommandPlugin(DigitizeCommand{}).
		WithResponse(202, `{"documentId":"648ea1c2-7dbe-42a8-b112-6474d07e61c1"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/00000000-0000-0000-0000-000000000000/digitization/result/648ea1c2-7dbe-42a8-b112-6474d07e61c1?api-version=1", 200, `{"status":"Done"}`).
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
	path := createFile(t)
	writeFile(path, []byte("hello-world"))

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
		WithCommandPlugin(DigitizeCommand{}).
		WithResponse(202, `{"documentId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", 200, `{"status":"Done"}`).
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
	path := createFile(t)
	writeFile(path, []byte("hello-world"))

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
		WithCommandPlugin(DigitizeCommand{}).
		WithResponse(202, `{"documentId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", 200, `{"pages":[],"status":"Done"}`).
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
	stdIn.Write([]byte("hello-world"))
	context := test.NewContextBuilder().
		WithDefinition("du", definition).
		WithConfig(config).
		WithCommandPlugin(DigitizeCommand{}).
		WithStdIn(stdIn).
		WithResponse(202, `{"documentId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", 200, `{"status":"Done"}`).
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
	path := createFile(t)
	writeFile(path, make([]byte, 10*1024*1024))

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
		WithCommandPlugin(DigitizeCommand{}).
		WithResponse(202, `{"documentId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/framework/projects/1234/digitization/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", 200, `{"status":"Done"}`).
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

func createFile(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "uipath-test")
	if err != nil {
		t.Fatal(err)
	}
	defer tempFile.Close()
	t.Cleanup(func() {
		err := os.Remove(tempFile.Name())
		if err != nil {
			t.Fatal(err)
		}
	})
	return tempFile.Name()
}

func writeFile(name string, data []byte) {
	err := os.WriteFile(name, data, 0600)
	if err != nil {
		panic(fmt.Errorf("Error writing file '%s': %w", name, err))
	}
}
