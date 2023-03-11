package digitzer

import (
	"bytes"
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

	result := test.RunCli([]string{"du", "digitization", "digitize"}, context)

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

	result := test.RunCli([]string{"du", "digitization", "digitize", "--file", "does-not-exist"}, context)

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

	result := test.RunCli([]string{"du", "digitization", "digitize", "--file", "hello-world"}, context)

	if !strings.Contains(result.StdErr, "Organization is not set") {
		t.Errorf("Expected stderr to show that organization parameter is missing, but got: %v", result.StdErr)
	}
}

func TestDigitizeWithFailedResponseReturnsError(t *testing.T) {
	path := createFile(t)
	os.WriteFile(path, []byte("hello-world"), 0644)

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

	result := test.RunCli([]string{"du", "digitization", "digitize", "--file", path}, context)

	if !strings.Contains(result.StdErr, "Digitizer returned status code '400' and body 'validation error'") {
		t.Errorf("Expected stderr to show that digitizer call failed, but got: %v", result.StdErr)
	}
}

func TestDigitizeSuccessfully(t *testing.T) {
	path := createFile(t)
	os.WriteFile(path, []byte("hello-world"), 0644)

	config := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	definition := `
servers:
- url: https://cloud.uipath.com/{organization}/{tenant}/du_/api/digitizer
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
		WithResponse(202, `{"operationId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/digitizer/digitize/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", 200, `{"status":"Done"}`).
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

func TestDigitizeSuccessfullyWithDebugFlag(t *testing.T) {
	path := createFile(t)
	os.WriteFile(path, []byte("hello-world"), 0644)

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
		WithResponse(202, `{"operationId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/digitizer/digitize/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", 200, `{"status":"Done"}`).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--file", path, "--debug"}, context)

	if !strings.Contains(result.StdOut, "/digitize/start") {
		t.Errorf("Expected stdout to show the start digitize operation, but got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "/digitize/result/eb80e441-05de-4a13-9aaa-f65b1babba05") {
		t.Errorf("Expected stdout to show the get digitize result operation, but got: %v", result.StdOut)
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
		WithResponse(202, `{"operationId":"eb80e441-05de-4a13-9aaa-f65b1babba05"}`).
		WithUrlResponse("/my-org/my-tenant/du_/api/digitizer/digitize/result/eb80e441-05de-4a13-9aaa-f65b1babba05?api-version=1", 200, `{"status":"Done"}`).
		Build()

	result := test.RunCli([]string{"du", "digitization", "digitize", "--content-type", "application/pdf"}, context)

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
	t.Cleanup(func() { os.Remove(tempFile.Name()) })
	return tempFile.Name()
}
