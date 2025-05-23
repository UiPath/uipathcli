package test

import (
	"net/http"
	"strings"
	"testing"
)

func TestOrganizationConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    organization: my-org
`
	definition := `
paths:
  "{organization}/ping":
    get:
      operationId: ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	if result.RequestUrl != "/my-org/ping" {
		t.Errorf("Did not set organization from config, got: %v", result.RequestUrl)
	}
}

func TestMissingOrganizationConfigShowsError(t *testing.T) {
	config := `
profiles:
  - name: default
`
	definition := `
paths:
  "{organization}/ping":
    get:
      operationId: ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	if !strings.HasPrefix(result.StdErr, "Missing organization parameter!") {
		t.Errorf("Did not show organization configuration error, got: %v", result.StdErr)
	}
}

func TestTenantConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    organization: my-org
    tenant: my-tenant
`
	definition := `
paths:
  "{organization}/{tenant}/ping":
    get:
      operationId: ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	if result.RequestUrl != "/my-org/my-tenant/ping" {
		t.Errorf("Did not set tenant from config, got: %v", result.RequestUrl)
	}
}

func TestOrgTenantServerUrl(t *testing.T) {
	config := `
profiles:
  - name: default
    organization: my-org
    tenant: my-tenant
`
	definition := `
servers:
  - url: https://cloud.uipath.com/{organization}/{tenant}/
paths:
  "/ping":
    get:
      operationId: ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	if result.RequestUrl != "/my-org/my-tenant/ping" {
		t.Errorf("Did not set tenant from config, got: %v", result.RequestUrl)
	}
}

func TestMissingTenantConfigShowsError(t *testing.T) {
	config := `
profiles:
  - name: default
    organization: my-org
`
	definition := `
paths:
  "{organization}/{tenant}/ping":
    get:
      operationId: ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	if !strings.HasPrefix(result.StdErr, "Missing tenant parameter!") {
		t.Errorf("Did not show tenant configuration error, got: %v", result.StdErr)
	}
}

func TestPathParameterConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    parameter:
      id: my-id
`
	definition := `
paths:
  /ping/{id}:
    get:
      summary: Ping operation
      operationId: ping
      parameters:
      - name: id
        in: path
        required: true
        description: The id
        schema:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	expected := "my-id"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Did not set correct custom path from config, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestQueryParameterConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    parameter:
      filter: my-filter
`
	definition := `
paths:
  /ping:
    get:
      summary: Ping operation
      parameters:
      - name: filter
        in: query
        required: true
        description: The filter
        schema:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	expected := "?filter=my-filter"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Did not set correct custom query string from config, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestHeaderParameterConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    parameter:
      x-uipath-test: abc
`
	definition := `
paths:
  /ping:
    get:
      summary: Ping operation
      parameters:
      - name: x-uipath-test
        in: header
        required: true
        description: The custom header
        schema:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	value := result.RequestHeader["x-uipath-test"]
	expected := "abc"
	if value != expected {
		t.Errorf("Did not set correct custom header from config, expected: %v, got: %v", expected, value)
	}
}

func TestCliParameterTakesPrecedenceOverConfigParameter(t *testing.T) {
	config := `
profiles:
  - name: default
    parameter:
      id: my-config-id
`
	definition := `
paths:
  /ping/{id}:
    get:
      summary: Ping operation
      operationId: ping
      parameters:
      - name: id
        in: path
        required: true
        description: The id
        schema:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping", "--id", "my-id"}, context)

	expected := "my-id"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Did not set correct parameter, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestConfigParameterTakesPrecedenceOverDefaultValue(t *testing.T) {
	config := `
profiles:
  - name: default
    parameter:
      test-filter: my-filter
`
	definition := `
paths:
  /ping/{TestFilter}:
    get:
      summary: Ping operation
      operationId: ping
      parameters:
      - name: TestFilter
        in: path
        required: true
        default: my-default-filter
        description: The test filter
        schema:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	expected := "my-filter"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Did not set correct parameter, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestConfigParameterNotPassedWhenNotDefined(t *testing.T) {
	config := `
profiles:
  - name: default
    parameter:
      x-uipath-test: abc
`
	definition := `
paths:
  /ping:
    get:
      summary: Ping operation
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	value := result.RequestHeader["x-uipath-test"]
	if value != "" {
		t.Errorf("Header parameter should not be sent, but got: %v", value)
	}
}

func TestCustomProfile(t *testing.T) {
	config := `
profiles:
  - name: default
    header:
      x-uipath-test: abc
  - name: myprofile
    header:
      x-uipath-test: 1234
`
	definition := `
paths:
  /ping:
    get:
      summary: Ping operation
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping", "--profile", "myprofile"}, context)

	value := result.RequestHeader["x-uipath-test"]
	expected := "1234"
	if value != expected {
		t.Errorf("Did not set load correct config profile, expected: %v, got '%v' for the header settings", expected, value)
	}
}

func TestInvalidProfileShowsError(t *testing.T) {
	config := `
profiles:
  - name: my-profile
    header:
      x-uipath-test: abc   
  `
	definition := `
paths:
  /ping:
    get:
      summary: test route
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		Build()

	result := RunCli([]string{"myservice", "get-ping", "--profile", "INVALID"}, context)

	expected := "Could not find profile 'INVALID'"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("stderr does not contain missing profile error, expected: %v, got: %v", expected, result.StdErr)
	}
}

func TestRequiredPathParameterFromConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    parameter:
      id: abc
`
	definition := `
paths:
  /ping/{id}:
    get:
      summary: Simple ping
      operationId: ping
      parameters:
      - name: id
        in: path
        required: true
        description: The id
        schema:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	if result.StdErr != "" {
		t.Errorf("Should not require path parameter when provided by config, got %v", result.StdErr)
	}
}

func TestRequiredQueryParameterFromConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    parameter:
      id: abc
`
	definition := `
paths:
  /ping/{id}:
    get:
      summary: Simple ping
      operationId: ping
      parameters:
      - name: id
        in: query
        required: true
        description: The id
        schema:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	if result.StdErr != "" {
		t.Errorf("Should not require query parameter when provided by config, got %v", result.StdErr)
	}
}

func TestRequiredHeaderParameterFromConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    parameter:
      x-uipath-test: abc
`
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      parameters:
      - name: x-uipath-test
        in: header
        required: true
        description: Test value
        schema:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	if result.StdErr != "" {
		t.Errorf("Should not require header parameter when provided by config, got %v", result.StdErr)
	}
}

func TestAdditionalHeaderParameterFromConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    header:
      x-uipath-test: abc
`
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	header := result.RequestHeader["x-uipath-test"]
	if header != "abc" {
		t.Errorf("Should send additional header from config, but got %v", header)
	}
}

func TestOutputFromConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    output: text
`
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(http.StatusOK, `{"a":"foo","b":1.1}`).
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	expectedOutput := "foo\t1.1\n"
	if result.StdOut != expectedOutput {
		t.Errorf("Should output text format, got %v", result.StdOut)
	}
}
