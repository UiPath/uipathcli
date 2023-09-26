package test

import (
	"os"
	"strings"
	"testing"
)

func TestConfigSetUnknownKey(t *testing.T) {
	context := NewContextBuilder().
		Build()

	result := RunCli([]string{"config", "set", "--key", "unknown", "--value", "my-value"}, context)

	if result.StdErr != "Unknown config key 'unknown'\n" {
		t.Errorf("Expected unknown config key error, but got %v", result.StdErr)
	}
}

func TestConfigSetCreatesNewProfile(t *testing.T) {
	configFile := createFile(t)
	existingConfig := `profiles:
- name: default
  organization: initial-org
`
	context := NewContextBuilder().
		WithConfig(existingConfig).
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "organization", "--value", "my-org", "--profile", "new"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: initial-org
- name: new
  organization: my-org
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigSetUpdatesExistingProfile(t *testing.T) {
	configFile := createFile(t)
	existingConfig := `profiles:
- name: existing
  organization: initial-org
`
	context := NewContextBuilder().
		WithConfig(existingConfig).
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "organization", "--value", "updated-org", "--profile", "existing"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: existing
  organization: updated-org
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigSetOrganization(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "organization", "--value", "my-org"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  organization: my-org
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigSetTenant(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "tenant", "--value", "my-tenant"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  tenant: my-tenant
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigSetUri(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "uri", "--value", "https://alpha.uipath.com"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  uri: https://alpha.uipath.com
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigInvalidUri(t *testing.T) {
	context := NewContextBuilder().
		Build()

	result := RunCli([]string{"config", "set", "--key", "uri", "--value", "invalid uri\t"}, context)

	if !strings.HasPrefix(result.StdErr, "Invalid value for 'uri'") {
		t.Errorf("Expected invalid uri error, but got %v", result.StdErr)
	}
}

func TestConfigSetInsecure(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "insecure", "--value", "true"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  insecure: true
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigSetDebug(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "debug", "--value", "TRUE"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  debug: true
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigInvalidInsecure(t *testing.T) {
	context := NewContextBuilder().
		Build()

	result := RunCli([]string{"config", "set", "--key", "insecure", "--value", "invalid"}, context)

	if !strings.HasPrefix(result.StdErr, "Invalid value for 'insecure'") {
		t.Errorf("Expected invalid insecure error, but got %v", result.StdErr)
	}
}

func TestConfigInvalidDebug(t *testing.T) {
	context := NewContextBuilder().
		Build()

	result := RunCli([]string{"config", "set", "--key", "debug", "--value", "invalid"}, context)

	if !strings.HasPrefix(result.StdErr, "Invalid value for 'debug'") {
		t.Errorf("Expected invalid debug error, but got %v", result.StdErr)
	}
}

func TestConfigSetHeader(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "header.x-uipath-license", "--value", "my-api-key"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  header:
    x-uipath-license: my-api-key
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigSetParameter(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "parameter.org", "--value", "my-org"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  parameter:
    org: my-org
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigSetAuthGrantType(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "auth.grantType", "--value", "password"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  auth:
    grantType: password
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigSetAuthScopes(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "auth.scopes", "--value", "MyScope"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  auth:
    scopes: MyScope
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigSetAuthProperties(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "auth.properties.acr_values", "--value", "tenant:host"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  auth:
    properties:
      acr_values: tenant:host
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}

func TestConfigSetVersion(t *testing.T) {
	configFile := createFile(t)
	context := NewContextBuilder().
		WithConfigFile(configFile).
		Build()

	RunCli([]string{"config", "set", "--key", "version", "--value", "22.10"}, context)

	config, err := os.ReadFile(configFile)
	if err != nil {
		t.Errorf("Config file does not exist: %v", err)
	}
	expectedConfig := `profiles:
- name: default
  version: "22.10"
`
	if string(config) != expectedConfig {
		t.Errorf("Expected generated config %v, but got %v", expectedConfig, string(config))
	}
}
