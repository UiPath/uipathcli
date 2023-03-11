package test

import (
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
		WithResponse(200, "").
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
		WithResponse(200, "").
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
		WithResponse(200, "").
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
		WithResponse(200, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	if !strings.HasPrefix(result.StdErr, "Missing tenant parameter!") {
		t.Errorf("Did not show tenant configuration error, got: %v", result.StdErr)
	}
}

func TestPathConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    path:
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
		WithResponse(200, "").
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	expected := "my-id"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Did not set correct custom path from config, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestQueryConfig(t *testing.T) {
	config := `
profiles:
  - name: default
    query:
      filter: my-filter
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
		WithResponse(200, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	expected := "?filter=my-filter"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Did not set correct custom query string from config, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestHeaderConfig(t *testing.T) {
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
      summary: Ping operation
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithConfig(config).
		WithResponse(200, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	value := result.RequestHeader["x-uipath-test"]
	expected := "abc"
	if value != expected {
		t.Errorf("Did not set correct custom header from config, expected: %v, got: %v", expected, value)
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
		WithResponse(200, "").
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
    path:
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
		WithResponse(200, "").
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
    query:
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
		WithResponse(200, "").
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
    header:
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
		WithResponse(200, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	if result.StdErr != "" {
		t.Errorf("Should not require header parameter when provided by config, got %v", result.StdErr)
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
		WithResponse(200, `{"a":"foo","b":1.1}`).
		Build()

	result := RunCli([]string{"myservice", "ping"}, context)

	expectedOutput := "foo\t1.1\n"
	if result.StdOut != expectedOutput {
		t.Errorf("Should output text format, got %v", result.StdOut)
	}
}
