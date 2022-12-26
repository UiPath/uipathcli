package commandline

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

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

	result := runCli([]string{"myservice", "ping"}, context)

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

	result := runCli([]string{"myservice", "get-ping"}, context)

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

	result := runCli([]string{"myservice", "get-ping"}, context)

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

	result := runCli([]string{"myservice", "get-ping", "--profile", "myprofile"}, context)

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

	result := runCli([]string{"myservice", "get-ping", "--profile", "INVALID"}, context)

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

	result := runCli([]string{"myservice", "ping"}, context)

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

	result := runCli([]string{"myservice", "ping"}, context)

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

	result := runCli([]string{"myservice", "get-ping"}, context)

	if result.StdErr != "" {
		t.Errorf("Should not require header parameter when provided by config, got %v", result.StdErr)
	}
}

func TestConfiguresCredentialsAuth(t *testing.T) {
	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("client-id\nclient-secret\nmy-org\nmy-tenant\n"))
	context := NewContextBuilder().
		WithStdIn(stdIn).
		Build()

	result := runCli([]string{"config"}, context)

	config, err := os.ReadFile(result.ConfigFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
  auth:
    clientId: client-id
    clientSecret: client-secret
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfiguresLoginAuth(t *testing.T) {
	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("ffe5141f-60fc-4fb9-8717-3969f303aedf\nhttp://localhost:27100\nOR.Users\nmy-org\nmy-tenant\n"))
	context := NewContextBuilder().
		WithStdIn(stdIn).
		Build()

	result := runCli([]string{"config", "--auth", "login"}, context)

	config, err := os.ReadFile(result.ConfigFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
  auth:
    clientId: ffe5141f-60fc-4fb9-8717-3969f303aedf
    redirectUri: http://localhost:27100
    scopes: OR.Users
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfiguresPatAuth(t *testing.T) {
	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("rt_mypersonalaccesstoken\nmy-org\nmy-tenant\n"))
	context := NewContextBuilder().
		WithStdIn(stdIn).
		Build()

	result := runCli([]string{"config", "--auth", "pat"}, context)

	config, err := os.ReadFile(result.ConfigFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
  auth:
    pat: rt_mypersonalaccesstoken
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfiguresPatAuthDoesNotChangeExistingConfigValues(t *testing.T) {
	config := `
profiles:
- name: default
  header:
    x-uipath-test: abc
`

	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("rt_mypersonalaccesstoken\nmy-org\nmy-tenant\n"))

	context := NewContextBuilder().
		WithConfig(config).
		WithStdIn(stdIn).
		Build()
	result := runCli([]string{"config", "--auth", "pat"}, context)

	updatedConfig, err := os.ReadFile(result.ConfigFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
  header:
    x-uipath-test: abc
  auth:
    pat: rt_mypersonalaccesstoken
`
	if string(updatedConfig) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(updatedConfig))
	}
}

func TestReconfiguresPatAuth(t *testing.T) {
	config := `
profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
  auth:
    pat: rt_mypersonalaccesstoken
`

	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("updated-token\nmy-updated-org\nmy-updated-tenant\n"))

	context := NewContextBuilder().
		WithConfig(config).
		WithStdIn(stdIn).
		Build()
	result := runCli([]string{"config", "--auth", "pat"}, context)

	updatedConfig, err := os.ReadFile(result.ConfigFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  path:
    organization: my-updated-org
    tenant: my-updated-tenant
  auth:
    pat: updated-token
`
	if string(updatedConfig) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(updatedConfig))
	}
}

func TestReconfiguresPatAuthPartially(t *testing.T) {
	config := `
profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
  auth:
    pat: rt_mypersonalaccesstoken
`

	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("\nmy-updated-org\n\n"))

	context := NewContextBuilder().
		WithConfig(config).
		WithStdIn(stdIn).
		Build()
	result := runCli([]string{"config", "--auth", "pat"}, context)

	updatedConfig, err := os.ReadFile(result.ConfigFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  path:
    organization: my-updated-org
    tenant: my-tenant
  auth:
    pat: rt_mypersonalaccesstoken
`
	if string(updatedConfig) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(updatedConfig))
	}
}

func TestConfiguresNewProfile(t *testing.T) {
	config := `
profiles:
- name: default
  header:
    x-uipath-test: abc
`

	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("rt_mypersonalaccesstoken\nmy-org\nmy-tenant\n"))

	context := NewContextBuilder().
		WithConfig(config).
		WithStdIn(stdIn).
		Build()
	result := runCli([]string{"config", "--auth", "pat", "--profile", "pat"}, context)

	updatedConfig, err := os.ReadFile(result.ConfigFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  header:
    x-uipath-test: abc
- name: pat
  path:
    organization: my-org
    tenant: my-tenant
  auth:
    pat: rt_mypersonalaccesstoken
`
	if string(updatedConfig) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(updatedConfig))
	}
}

func TestReconfiguresExistingProfile(t *testing.T) {
	config := `
profiles:
- name: default
  header:
    x-uipath-test: abc
- name: pat
  path:
    organization: my-org
  auth:
    pat: rt_mypersonalaccesstoken
`

	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("my-new-token\n\n\n"))

	context := NewContextBuilder().
		WithConfig(config).
		WithStdIn(stdIn).
		Build()
	result := runCli([]string{"config", "--auth", "pat", "--profile", "pat"}, context)

	updatedConfig, err := os.ReadFile(result.ConfigFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  header:
    x-uipath-test: abc
- name: pat
  path:
    organization: my-org
  auth:
    pat: my-new-token
`
	if string(updatedConfig) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(updatedConfig))
	}
}

func TestCredentialsAuthOutputNotSet(t *testing.T) {
	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("\n\n\n\n"))

	context := NewContextBuilder().
		WithStdIn(stdIn).
		Build()
	result := runCli([]string{"config"}, context)

	expectedOutput := `Enter client id [not set]: Enter client secret [not set]: Enter organization [not set]: Enter tenant [not set]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}

func TestCredentialsAuthMasksSecrets(t *testing.T) {
	config := `
profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
  auth:
    clientId: 433d7778-8440-4e74-81f0-d88351bde871
    clientSecret: UaX#Fen)8mvifo
`

	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("\n\n\n\n"))

	context := NewContextBuilder().
		WithConfig(config).
		WithStdIn(stdIn).
		Build()
	result := runCli([]string{"config"}, context)

	expectedOutput := `Enter client id [*******e871]: Enter client secret [*******vifo]: Enter organization [my-org]: Enter tenant [my-tenant]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}

func TestCredentialsAuthMasksShortSecretsCompletely(t *testing.T) {
	config := `
profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
  auth:
    clientId: very
    clientSecret: short
`

	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("\n\n\n\n"))

	context := NewContextBuilder().
		WithConfig(config).
		WithStdIn(stdIn).
		Build()
	result := runCli([]string{"config"}, context)

	expectedOutput := `Enter client id [*******]: Enter client secret [*******]: Enter organization [my-org]: Enter tenant [my-tenant]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}

func TestPatAuthMasksSecrets(t *testing.T) {
	config := `
profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
  auth:
    pat: f73b2edb-b37b-4426-8cc8-e7f98b09a827
`

	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("\n\n\n\n"))

	context := NewContextBuilder().
		WithConfig(config).
		WithStdIn(stdIn).
		Build()
	result := runCli([]string{"config", "--auth", "pat"}, context)

	expectedOutput := `Enter personal access token [*******a827]: Enter organization [my-org]: Enter tenant [my-tenant]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}

func TestLoginAuthMasksSecrets(t *testing.T) {
	config := `
profiles:
- name: default
  path:
    organization: my-org
    tenant: my-tenant
  auth:
    clientId: 891979c1-68e2-46bb-9016-e5f2241fdd35
    redirectUri: http://localhost:27100
    scopes: OR.Users.Read OR.Users.Write
`

	stdIn := bytes.Buffer{}
	stdIn.Write([]byte("\n\n\n\n"))

	context := NewContextBuilder().
		WithConfig(config).
		WithStdIn(stdIn).
		Build()
	result := runCli([]string{"config", "--auth", "login"}, context)

	expectedOutput := `Enter client id [*******dd35]: Enter redirect uri [http://localhost:27100]: Enter scopes [OR.Users.Read OR.Users.Write]: Enter organization [my-org]: Enter tenant [my-tenant]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}
