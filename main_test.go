package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainReadsDefinitions(t *testing.T) {
	config := createFile(t, "config", "")
	definition := createFile(t, "service-a.yaml", "")

	t.Setenv("UIPATH_CONFIGURATION_PATH", config)
	t.Setenv("UIPATH_DEFINITIONS_PATH", filepath.Dir(definition))

	os.Args = []string{"uipath", "--help"}
	output := captureOutput(t, func() {
		main()
	})

	expected := `service-a`
	if !strings.Contains(output, expected) {
		t.Errorf("Expected %s in output, but got: %v", expected, output)
	}
}

func TestHelpReadsDefinitions(t *testing.T) {
	config := createFile(t, "config", "")
	definition := createFile(t, "service-a.yaml", "")

	t.Setenv("UIPATH_CONFIGURATION_PATH", config)
	t.Setenv("UIPATH_DEFINITIONS_PATH", filepath.Dir(definition))

	os.Args = []string{"uipath", "-h"}
	output := captureOutput(t, func() {
		main()
	})

	expected := `service-a`
	if !strings.Contains(output, expected) {
		t.Errorf("Expected %s in output, but got: %v", expected, output)
	}
}

func TestMainParsesDefinition(t *testing.T) {
	config := createFile(t, "config", "")
	definition := createFile(t, "service-a.yaml", `
paths:
  /ping:
    get:
      summary: This is a simple get operation
      operationId: ping
`)

	t.Setenv("UIPATH_CONFIGURATION_PATH", config)
	t.Setenv("UIPATH_DEFINITIONS_PATH", filepath.Dir(definition))

	os.Args = []string{"uipath", "service-a", "--help"}
	output := captureOutput(t, func() {
		main()
	})

	expected := `ping`
	if !strings.Contains(output, expected) {
		t.Errorf("Expected operation name %s in output, but got: %v", expected, output)
	}
	expected = `This is a simple get operation`
	if !strings.Contains(output, expected) {
		t.Errorf("Expected description %s in output, but got: %v", expected, output)
	}
}

func TestMainCallsService(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/identity_/connect/token" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1234}`))
	}))
	defer srv.Close()

	config := createFile(t, "config", `
profiles:
- name: default
  uri: `+srv.URL+`
  organization: my-org
  tenant: defaulttenant
  auth:
    clientId: 71b784bc-3f7b-4e5a-a731-db25bb829025
    clientSecret: NGI&4b(chsHcsX^C
`)

	definition := createFile(t, "service-a.yaml", `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`)

	t.Setenv("UIPATH_CONFIGURATION_PATH", config)
	t.Setenv("UIPATH_DEFINITIONS_PATH", filepath.Dir(definition))

	os.Args = []string{"uipath", "service-a", "ping"}
	output := captureOutput(t, func() {
		main()
	})

	expected := `{
  "id": 1234
}`
	if !strings.Contains(output, expected) {
		t.Errorf("Expected %s in output, but got: %v", expected, output)
	}
}

func TestMainAutocompletesCommand(t *testing.T) {
	config := createFile(t, "config", "")
	definition := createFile(t, "service-a.yaml", `
paths:
  /ping:
    get:
      summary: This is a simple get operation
      operationId: ping
`)

	t.Setenv("UIPATH_CONFIGURATION_PATH", config)
	t.Setenv("UIPATH_DEFINITIONS_PATH", filepath.Dir(definition))

	os.Args = []string{"uipath", "autocomplete", "complete", "--command", "upathcli service-a p"}
	output := captureOutput(t, func() {
		main()
	})

	expected := `ping
`
	if output != expected {
		t.Errorf("Expected operation name %s in autocomplete output, but got: %v", expected, output)
	}
}

func TestMainParsesBuiltInDefinitions(t *testing.T) {
	t.Run("du-framework", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "du discovery", "projects") })
	t.Run("orchestrator", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "orchestrator users", "get-by-id") })
	t.Run("orchestrator", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "orchestrator", "assets") })
}

func MainParsesBuiltInDefinitions(t *testing.T, command string, expected string) {
	definitionDir, _ := os.Getwd()
	t.Setenv("UIPATH_DEFINITIONS_PATH", filepath.Join(definitionDir, "definitions/"))

	os.Args = strings.Split("uipath "+command, " ")
	output := captureOutput(t, func() {
		main()
	})

	if !strings.Contains(output, expected) {
		t.Errorf("Expected %s in output, but got: %v", expected, output)
	}
}

func captureOutput(t *testing.T, runnable func()) string {
	realStdout := os.Stdout
	reader, fakeStdout, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { os.Stdout = realStdout }()
	os.Stdout = fakeStdout

	output := make(chan []byte)
	go func(reader *os.File) {
		defer func() { _ = reader.Close() }()
		data, err := io.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		output <- data
	}(reader)

	runnable()

	err = fakeStdout.Close()
	if err != nil {
		t.Fatal(err)
	}
	return string(<-output)
}

func createFile(t *testing.T, name string, content string) string {
	path := filepath.Join(t.TempDir(), name)
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatal(fmt.Errorf("Error writing file '%s': %w", path, err))
	}
	return path
}
