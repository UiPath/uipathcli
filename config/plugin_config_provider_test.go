package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestEmptyPluginConfigWhenPluginFileNotFound(t *testing.T) {
	configProvider := NewPluginConfigProvider(NewPluginConfigFileStore("no-plugin-file"))

	err := configProvider.Load()
	config := configProvider.Config()

	if err != nil {
		t.Errorf("Loading plugin config should not return an error, but got: %v", err)
	}
	if len(config.Authenticators) != 0 {
		t.Errorf("Plugin config should not contain any authenticators, but got: %v", config.Authenticators)
	}
}

func TestErrorOnPluginFileParsingError(t *testing.T) {
	file := createFile(t)
	writeFile(file, []byte("INVALID CONTENT"))
	configProvider := NewPluginConfigProvider(NewPluginConfigFileStore(file))

	err := configProvider.Load()
	if !strings.HasPrefix(err.Error(), "yaml: unmarshal errors") {
		t.Errorf("Should show plugin file parsing error, but got: %v", err)
	}
}

func TestPluginFileSuccessfullyParsed(t *testing.T) {
	plugin := `
authenticators:
  - name: kubernetes
    path: ./uipathcli-authenticator-k8s
`
	file := createFile(t)
	writeFile(file, []byte(plugin))
	configProvider := NewPluginConfigProvider(NewPluginConfigFileStore(file))

	err := configProvider.Load()
	if err != nil {
		t.Errorf("Should load plugin configuration file, but got: %v", err)
	}
}

func TestPluginConfigValid(t *testing.T) {
	plugin := `
authenticators:
  - name: kubernetes
    path: ./uipathcli-authenticator-k8s
`
	file := createFile(t)
	writeFile(file, []byte(plugin))
	configProvider := NewPluginConfigProvider(NewPluginConfigFileStore(file))

	err := configProvider.Load()
	config := configProvider.Config()
	if err != nil {
		t.Errorf("Loading plugin config should not return an error, but got: %v", err)
	}
	if config.Authenticators[0].Name != "kubernetes" {
		t.Errorf("Should parse authenticator name, but got: %v", config)
	}
	if config.Authenticators[0].Path != "./uipathcli-authenticator-k8s" {
		t.Errorf("Should parse authenticator path, but got: %v", config)
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

func writeFile(name string, data []byte) {
	err := os.WriteFile(name, data, 0600)
	if err != nil {
		panic(fmt.Errorf("Error writing file '%s': %w", name, err))
	}
}
