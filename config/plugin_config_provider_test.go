package config

import (
	"os"
	"strings"
	"testing"
)

func TestEmptyPluginConfigWhenPluginFileNotFound(t *testing.T) {
	configProvider := NewPluginConfigProvider(*NewPluginConfigStore("no-plugin-file"))

	configProvider.Load()
	config := configProvider.Config()
	if len(config.Authenticators) != 0 {
		t.Errorf("Plugin config should not contain any authenticators, but got: %v", config.Authenticators)
	}
}

func TestErrorOnPluginFileParsingError(t *testing.T) {
	file := createFile(t)
	os.WriteFile(file, []byte("INVALID CONTENT"), 0600)
	configProvider := NewPluginConfigProvider(*NewPluginConfigStore(file))

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
	os.WriteFile(file, []byte(plugin), 0600)
	configProvider := NewPluginConfigProvider(*NewPluginConfigStore(file))

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
	os.WriteFile(file, []byte(plugin), 0600)
	configProvider := NewPluginConfigProvider(*NewPluginConfigStore(file))

	configProvider.Load()
	config := configProvider.Config()
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
