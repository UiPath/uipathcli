package test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestConfigCommandIsShown(t *testing.T) {
	context := NewContextBuilder().
		Build()

	result := RunCli([]string{"--help"}, context)

	if !strings.Contains(result.StdOut, "config") {
		t.Errorf("Expected config command to be shown, but got %v", result.StdOut)
	}
}

func TestConfigCommandDescriptionIsShown(t *testing.T) {
	context := NewContextBuilder().
		Build()

	result := RunCli([]string{"config", "--help"}, context)

	if !strings.Contains(result.StdOut, "Interactive command to configure the CLI") {
		t.Errorf("Expected config command description to be shown, but got %v", result.StdOut)
	}
}

func TestConfiguresCredentialsAuth(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("client-id\nclient-secret\n1\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		WithTokenResponse(http.StatusOK, `{"access_token": "`+createToken("d653ed49-3483-4118-9fa0-cbc7c9651c06")+`", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		WithUrlResponse(
			"/d653ed49-3483-4118-9fa0-cbc7c9651c06/organization_/api/organization/d653ed49-3483-4118-9fa0-cbc7c9651c06/AllInfo",
			http.StatusOK,
			`{
			   "organization": { "id":"d653ed49-3483-4118-9fa0-cbc7c9651c06", "logicalName":"my-org" },
			   "tenants": [
			     { "id":"0431e765-42bb-417b-bb13-77e256f4c898", "name":"my-tenant" },
			     { "id":"be6b13fb-2a3e-4ed1-aa9f-e163fb664d8b", "name":"my-second-tenant" }
			   ]
			 }`,
		).
		Build()

	RunCli([]string{"config", "--auth", "credentials"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
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
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("ffe5141f-60fc-4fb9-8717-3969f303aedf\n\nhttp://localhost:27100\nOR.Users\n1\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		WithOAuthTokenResponse(http.StatusOK, `{"access_token": "`+createToken("d653ed49-3483-4118-9fa0-cbc7c9651c06")+`", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		WithUrlResponse(
			"/d653ed49-3483-4118-9fa0-cbc7c9651c06/organization_/api/organization/d653ed49-3483-4118-9fa0-cbc7c9651c06/AllInfo",
			http.StatusOK,
			`{
			   "organization": { "id":"d653ed49-3483-4118-9fa0-cbc7c9651c06", "logicalName":"my-org" },
			   "tenants": [
			     { "id":"0431e765-42bb-417b-bb13-77e256f4c898", "name":"my-tenant" },
			     { "id":"be6b13fb-2a3e-4ed1-aa9f-e163fb664d8b", "name":"my-second-tenant" }
			   ]
			 }`,
		).
		Build()

	RunCli([]string{"config", "--auth", "login"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
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

func TestConfiguresLoginConfidentialAuth(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("ffe5141f-60fc-4fb9-8717-3969f303aedf\nmy-secret\nhttp://localhost:27100\nOR.Users\n1\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		WithOAuthTokenResponse(http.StatusOK, `{"access_token": "`+createToken("d653ed49-3483-4118-9fa0-cbc7c9651c06")+`", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		WithUrlResponse(
			"/d653ed49-3483-4118-9fa0-cbc7c9651c06/organization_/api/organization/d653ed49-3483-4118-9fa0-cbc7c9651c06/AllInfo",
			http.StatusOK,
			`{
			   "organization": { "id":"d653ed49-3483-4118-9fa0-cbc7c9651c06", "logicalName":"my-org" },
			   "tenants": [
			     { "id":"0431e765-42bb-417b-bb13-77e256f4c898", "name":"my-tenant" },
			     { "id":"be6b13fb-2a3e-4ed1-aa9f-e163fb664d8b", "name":"my-second-tenant" }
			   ]
			 }`,
		).
		Build()

	RunCli([]string{"config", "--auth", "login"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: ffe5141f-60fc-4fb9-8717-3969f303aedf
    clientSecret: my-secret
    redirectUri: http://localhost:27100
    scopes: OR.Users
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfiguresPatAuth(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("rt_mypersonalaccesstoken\nmy-org\nmy-tenant\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "--auth", "pat"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
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
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  header:
    x-uipath-test: abc
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("rt_mypersonalaccesstoken\nmy-org\nmy-tenant\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	RunCli([]string{"config", "--auth", "pat"}, context)

	updatedConfig, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
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

func TestReconfiguresExistingLoginConfidentialAuthAsNonConfidential(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: 06572c32-8ebe-4e0a-b067-844bc3818d58
    clientSecret: my-secret
    redirectUri: http://localhost:27100
    scopes: OR.Default
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n \n\n\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	RunCli([]string{"config", "--auth", "login"}, context)

	updatedConfig, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: 06572c32-8ebe-4e0a-b067-844bc3818d58
    clientSecret: ""
    redirectUri: http://localhost:27100
    scopes: OR.Default
`
	if string(updatedConfig) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(updatedConfig))
	}
}

func TestReconfiguresExistingLoginConfidentialAuthAsCredentials(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: 06572c32-8ebe-4e0a-b067-844bc3818d58
    clientSecret: my-secret
    redirectUri: http://localhost:27100
    scopes: OR.Default
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n1\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		WithTokenResponse(http.StatusOK, `{"access_token": "`+createToken("d653ed49-3483-4118-9fa0-cbc7c9651c06")+`", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		WithUrlResponse(
			"/d653ed49-3483-4118-9fa0-cbc7c9651c06/organization_/api/organization/d653ed49-3483-4118-9fa0-cbc7c9651c06/AllInfo",
			http.StatusOK,
			`{
			   "organization": { "id":"d653ed49-3483-4118-9fa0-cbc7c9651c06", "logicalName":"my-org" },
			   "tenants": [
			     { "id":"0431e765-42bb-417b-bb13-77e256f4c898","name":"my-tenant" },
			     { "id":"be6b13fb-2a3e-4ed1-aa9f-e163fb664d8b", "name":"my-second-tenant" }
			   ]
			 }`,
		).
		Build()
	RunCli([]string{"config", "--auth", "credentials"}, context)

	updatedConfig, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: 06572c32-8ebe-4e0a-b067-844bc3818d58
    clientSecret: my-secret
`
	if string(updatedConfig) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(updatedConfig))
	}
}

func TestReconfiguresPatAuth(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    pat: rt_mypersonalaccesstoken
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("updated-token\nmy-updated-org\nmy-updated-tenant\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	RunCli([]string{"config", "--auth", "pat"}, context)

	updatedConfig, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
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
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    pat: rt_mypersonalaccesstoken
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\nmy-updated-org\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	RunCli([]string{"config", "--auth", "pat"}, context)

	updatedConfig, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
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
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  header:
    x-uipath-test: abc
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("rt_mypersonalaccesstoken\nmy-org\nmy-tenant\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	RunCli([]string{"config", "--auth", "pat", "--profile", "pat"}, context)

	updatedConfig, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  header:
    x-uipath-test: abc
- name: pat
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
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  header:
    x-uipath-test: abc
- name: pat
  organization: my-org
  auth:
    pat: rt_mypersonalaccesstoken
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("my-new-token\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	RunCli([]string{"config", "--auth", "pat", "--profile", "pat"}, context)

	updatedConfig, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  header:
    x-uipath-test: abc
- name: pat
  organization: my-org
  auth:
    pat: my-new-token
`
	if string(updatedConfig) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(updatedConfig))
	}
}

func TestMultiAuthOutputNotSet(t *testing.T) {
	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n")

	context := NewContextBuilder().
		WithStdIn(stdIn).
		Build()
	result := RunCli([]string{"config"}, context)

	expectedOutput := `Authentication type [not set]:
  [1] credentials - Client Id and Client Secret
  [2] login - OAuth login using the browser
  [3] pat - Personal Access Token
Select: Enter organization [not set]: Enter tenant [not set]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt '%v', but got '%v'", expectedOutput, result.StdOut)
	}
}

func TestCredentialsAuthOutputNotSet(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n\n")

	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		Build()
	result := RunCli([]string{"config", "--auth", "credentials"}, context)

	expectedOutput := `Enter client id [not set]: Enter client secret [not set]: Enter organization [not set]: Enter tenant [not set]: Successfully configured uipath CLI
`
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}

func TestCredentialsAuthMasksSecrets(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: 433d7778-8440-4e74-81f0-d88351bde871
    clientSecret: UaX#Fen)8mvifo
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	result := RunCli([]string{"config", "--auth", "credentials"}, context)

	expectedOutput := `Enter client id [*******e871]: Enter client secret [*******vifo]: Enter organization [my-org]: Enter tenant [my-tenant]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}

func TestCredentialsAuthMasksShortSecretsCompletely(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: very
    clientSecret: short
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	result := RunCli([]string{"config", "--auth", "credentials"}, context)

	expectedOutput := `Enter client id [*******]: Enter client secret [*******]: Enter organization [my-org]: Enter tenant [my-tenant]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}

func TestPatAuthMasksSecrets(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    pat: f73b2edb-b37b-4426-8cc8-e7f98b09a827
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	result := RunCli([]string{"config", "--auth", "pat"}, context)

	expectedOutput := `Enter personal access token [*******a827]: Enter organization [my-org]: Enter tenant [my-tenant]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}

func TestLoginAuthMasksSecrets(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: 891979c1-68e2-46bb-9016-e5f2241fdd35
    redirectUri: http://localhost:27100
    scopes: OR.Users.Read OR.Users.Write
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	result := RunCli([]string{"config", "--auth", "login"}, context)

	expectedOutput := `Enter client id [*******dd35]: Enter client secret (only for confidential apps) [not set]: Enter redirect uri [http://localhost:27100]: Enter scopes [OR.Users.Read OR.Users.Write]: Enter organization [my-org]: Enter tenant [my-tenant]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}

func TestLoginConfidentialAuthMasksSecrets(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: 891979c1-68e2-46bb-9016-e5f2241fdd35
    clientSecret: my-secret
    redirectUri: http://localhost:27100
    scopes: OR.Users.Read OR.Users.Write
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	result := RunCli([]string{"config", "--auth", "login"}, context)

	expectedOutput := `Enter client id [*******dd35]: Enter client secret (only for confidential apps) [*******]: Enter redirect uri [http://localhost:27100]: Enter scopes [OR.Users.Read OR.Users.Write]: Enter organization [my-org]: Enter tenant [my-tenant]: `
	if result.StdOut != expectedOutput {
		t.Errorf("Expected prompt %v, but got %v", expectedOutput, result.StdOut)
	}
}

func TestConfigureMultiAuthCredentialsAuth(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("1\nclient-id\nclient-secret\nmy-org\nmy-tenant\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
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

func TestConfigureMultiAuthNonLoginConfidentialAuth(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("2\nffe5141f-60fc-4fb9-8717-3969f303aedf\n\nhttp://localhost:27100\nOR.Users\nmy-org\nmy-tenant\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
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

func TestConfigureMultiAuthLoginConfidentialAuth(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("2\nffe5141f-60fc-4fb9-8717-3969f303aedf\nmy-secret\nhttp://localhost:27100\nOR.Users\nmy-org\nmy-tenant\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: ffe5141f-60fc-4fb9-8717-3969f303aedf
    clientSecret: my-secret
    redirectUri: http://localhost:27100
    scopes: OR.Users
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigureMultiAuthPatAuth(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("3\nrt_mypersonalaccesstoken\nmy-org\nmy-tenant\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    pat: rt_mypersonalaccesstoken
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigureMultiAuthShowsExistingCredentialsAuth(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: my-client-id
    clientSecret: my-secret
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	result := RunCli([]string{"config"}, context)

	if !strings.Contains(result.StdOut, "Authentication type [credentials]:") {
		t.Errorf("Expected existing authentication type credentials, but got %v", result.StdOut)
	}
}

func TestConfigureMultiAuthShowsExistingLoginAuth(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: my-client-id
    redirectUri: http://localhost:12700
    scopes: OR.users
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	result := RunCli([]string{"config"}, context)

	if !strings.Contains(result.StdOut, "Authentication type [login]:") {
		t.Errorf("Expected existing authentication type login, but got %v", result.StdOut)
	}
}

func TestConfigureMultiAuthShowsExistingPatAuth(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    pat: my-pat
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	result := RunCli([]string{"config"}, context)

	if !strings.Contains(result.StdOut, "Authentication type [pat]:") {
		t.Errorf("Expected existing authentication type pat, but got %v", result.StdOut)
	}
}

func TestConfigureMultiAuthShowsNoAuthSet(t *testing.T) {
	configFile := TempFile(t)
	config := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\n\n\n")

	context := NewContextBuilder().
		WithConfig(config).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	result := RunCli([]string{"config"}, context)

	if !strings.Contains(result.StdOut, "Authentication type [not set]:") {
		t.Errorf("Expected no authentication type, but got %v", result.StdOut)
	}
}

func TestConfigureMultiAuthModifiesExistingPatAuth(t *testing.T) {
	configFile := TempFile(t)
	existingConfig := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    pat: f73b2edb-b37b-4426-8cc8-e7f98b09a827
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\nnew-pat\n\n\n")

	context := NewContextBuilder().
		WithConfig(existingConfig).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	RunCli([]string{"config"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    pat: new-pat
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigureMultiAuthModifiesExistingCredentialsAuth(t *testing.T) {
	configFile := TempFile(t)
	existingConfig := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: my-client-id
    clientSecret: my-client-secret
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\nnew-client-id\nnew-client-secret\n\n\n")

	context := NewContextBuilder().
		WithConfig(existingConfig).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	RunCli([]string{"config"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: new-client-id
    clientSecret: new-client-secret
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigureMultiAuthModifiesExistingLoginAuth(t *testing.T) {
	configFile := TempFile(t)
	existingConfig := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: ffe5141f-60fc-4fb9-8717-3969f303aedf
    redirectUri: http://localhost:27100
    scopes: OR.Users
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\nb2f0fa8a-8a79-4733-b810-fe9989e39334\n\nhttp://new-url:8080\nOR.Machines\n\n\n")

	context := NewContextBuilder().
		WithConfig(existingConfig).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	RunCli([]string{"config"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: b2f0fa8a-8a79-4733-b810-fe9989e39334
    redirectUri: http://new-url:8080
    scopes: OR.Machines
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigureMultiAuthModifiesExistingLoginConfidentialAuth(t *testing.T) {
	configFile := TempFile(t)
	existingConfig := `
profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: 6ff4a796-a938-4a75-82ab-e7e8c2577720
    clientSecret: my-secret
    redirectUri: http://localhost:27100
    scopes: OR.Default
`

	stdIn := bytes.Buffer{}
	stdIn.WriteString("\nadb7e1b3-6008-4f24-9ab4-4cac435987f8\nmy-updated-secret\nhttp://new-url:8080\nOR.Folders\n\n\n")

	context := NewContextBuilder().
		WithConfig(existingConfig).
		WithConfigFile(configFile).
		WithStdIn(stdIn).
		Build()
	RunCli([]string{"config"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
  tenant: my-tenant
  auth:
    clientId: adb7e1b3-6008-4f24-9ab4-4cac435987f8
    clientSecret: my-updated-secret
    redirectUri: http://new-url:8080
    scopes: OR.Folders
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestSelectsFromMultipleTenants(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("client-id\nclient-secret\n2\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		WithTokenResponse(http.StatusOK, `{"access_token": "`+createToken("d653ed49-3483-4118-9fa0-cbc7c9651c06")+`", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		WithUrlResponse(
			"/d653ed49-3483-4118-9fa0-cbc7c9651c06/organization_/api/organization/d653ed49-3483-4118-9fa0-cbc7c9651c06/AllInfo",
			http.StatusOK,
			`{
			   "organization": { "id": "d653ed49-3483-4118-9fa0-cbc7c9651c06", "logicalName":"my-org" },
			   "tenants": [
			     { "id":"ac4c87fa-d113-416e-8b11-49ec59c3df07", "name":"DefaultTenant" },
			     { "id":"be6b13fb-2a3e-4ed1-aa9f-e163fb664d8b", "name":"MySecondTenant" }
			   ]
			 }`,
		).
		Build()

	RunCli([]string{"config", "--auth", "credentials"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
  tenant: MySecondTenant
  auth:
    clientId: client-id
    clientSecret: client-secret
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestPreselectsSingleTenant(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("268f235c-ce53-4b69-a744-d8b89c1f9327\n#asfB#Q$fGRxF1iz)\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		WithTokenResponse(http.StatusOK, `{"access_token": "`+createToken("d653ed49-3483-4118-9fa0-cbc7c9651c06")+`", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		WithUrlResponse(
			"/d653ed49-3483-4118-9fa0-cbc7c9651c06/organization_/api/organization/d653ed49-3483-4118-9fa0-cbc7c9651c06/AllInfo",
			http.StatusOK,
			`{
			   "organization": { "id": "d653ed49-3483-4118-9fa0-cbc7c9651c06", "logicalName":"uiptestorg" },
			   "tenants": [
			     { "id":"ac4c87fa-d113-416e-8b11-49ec59c3df07", "name":"DefaultTenant" }
			   ]
			 }`,
		).
		Build()

	RunCli([]string{"config", "--auth", "credentials"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: uiptestorg
  tenant: DefaultTenant
  auth:
    clientId: 268f235c-ce53-4b69-a744-d8b89c1f9327
    clientSecret: '#asfB#Q$fGRxF1iz)'
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestSelectRepeatsWhenInvalidTenant(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("client-id\nclient-secret\ninvalid\n3\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		WithTokenResponse(http.StatusOK, `{"access_token": "`+createToken("d653ed49-3483-4118-9fa0-cbc7c9651c06")+`", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		WithUrlResponse(
			"/d653ed49-3483-4118-9fa0-cbc7c9651c06/organization_/api/organization/d653ed49-3483-4118-9fa0-cbc7c9651c06/AllInfo",
			http.StatusOK,
			`{
			   "organization": { "id": "d653ed49-3483-4118-9fa0-cbc7c9651c06", "logicalName":"my-org" },
			   "tenants": [
			     { "id":"ac4c87fa-d113-416e-8b11-49ec59c3df07", "name":"DefaultTenant" },
			     { "id":"be6b13fb-2a3e-4ed1-aa9f-e163fb664d8b", "name":"MySecondTenant" }
			   ]
			 }`,
		).
		Build()

	result := RunCli([]string{"config", "--auth", "credentials"}, context)

	if strings.Count(result.StdOut, "Tenant [not set]:") != 3 {
		t.Errorf("Expected repeating tenant prompt, but got %v", result.StdOut)
	}
}

func TestHandlesFailedGetTenantsCallAndFallsBackToEnteringOrgManually(t *testing.T) {
	configFile := TempFile(t)

	stdIn := bytes.Buffer{}
	stdIn.WriteString("client-id\nclient-secret\nmy-org\nmy-tenant\n")
	context := NewContextBuilder().
		WithStdIn(stdIn).
		WithConfigFile(configFile).
		WithTokenResponse(http.StatusOK, `{"access_token": "`+createToken("d653ed49-3483-4118-9fa0-cbc7c9651c06")+`", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		WithUrlResponse("/d653ed49-3483-4118-9fa0-cbc7c9651c06/organization_/api/organization/d653ed49-3483-4118-9fa0-cbc7c9651c06/AllInfo", http.StatusNotFound, "").
		Build()

	result := RunCli([]string{"config", "--auth", "credentials"}, context)

	if !strings.Contains(result.StdErr, `
Authentication Failed!
Organization Management Service returned status code '404' and body ''
Please make sure the credentials are correct or enter the organization and tenant information manually.`) {
		t.Errorf("Expected error message, but got %v", result.StdErr)
	}

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
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

func createToken(partId string) string {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJson, _ := json.Marshal(header)
	headerEnc := base64.StdEncoding.EncodeToString(headerJson)

	payload := map[string]string{
		"prt_id": partId,
	}
	payloadJson, _ := json.Marshal(payload)
	payloadEnc := base64.StdEncoding.EncodeToString(payloadJson)

	signatureEnc := base64.StdEncoding.EncodeToString([]byte("dummysignature"))

	return headerEnc + "." + payloadEnc + "." + signatureEnc
}
