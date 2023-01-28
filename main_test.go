package main

import (
	"encoding/hex"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestMainReadsDefinitions(t *testing.T) {
	config := createFile(t, ".uipathcli", "config")
	definition := createFile(t, "definitions", "service-a.yaml")

	t.Setenv("UIPATHCLI_CONFIGURATION_PATH", config)
	t.Setenv("UIPATHCLI_DEFINITIONS_PATH", filepath.Dir(definition))

	os.Args = []string{"uipathcli", "--help"}
	output := captureOutput(t, func() {
		main()
	})

	expected := `service-a`
	if !strings.Contains(output, expected) {
		t.Errorf("Expected %s in output, but got: %v", expected, output)
	}
}

func TestMainParsesDefinition(t *testing.T) {
	config := createFile(t, ".uipathcli", "config")
	definition := createFile(t, "definitions", "service-a.yaml")
	os.WriteFile(definition, []byte(`
paths:
  /ping:
    get:
      summary: This is a simple get operation
      operationId: ping
`), 0600)

	t.Setenv("UIPATHCLI_CONFIGURATION_PATH", config)
	t.Setenv("UIPATHCLI_DEFINITIONS_PATH", filepath.Dir(definition))

	os.Args = []string{"uipathcli", "service-a", "--help"}
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
			w.WriteHeader(200)
			w.Write([]byte(`{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`))
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(`{"id":1234}`))
	}))
	defer srv.Close()

	config := createFile(t, ".uipathcli", "config")
	os.WriteFile(config, []byte(`
profiles:
- name: default
  uri: `+srv.URL+`
  path:
    organization: my-org
    tenant: defaulttenant
  auth:
    clientId: 71b784bc-3f7b-4e5a-a731-db25bb829025
    clientSecret: NGI&4b(chsHcsX^C
`), 0600)

	definition := createFile(t, "definitions", "service-a.yaml")
	os.WriteFile(definition, []byte(`
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`), 0600)

	t.Setenv("UIPATHCLI_CONFIGURATION_PATH", config)
	t.Setenv("UIPATHCLI_DEFINITIONS_PATH", filepath.Dir(definition))

	os.Args = []string{"uipathcli", "service-a", "ping"}
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
	config := createFile(t, ".uipathcli", "config")
	definition := createFile(t, "definitions", "service-a.yaml")
	os.WriteFile(definition, []byte(`
paths:
  /ping:
    get:
      summary: This is a simple get operation
      operationId: ping
`), 0600)

	t.Setenv("UIPATHCLI_CONFIGURATION_PATH", config)
	t.Setenv("UIPATHCLI_DEFINITIONS_PATH", filepath.Dir(definition))

	os.Args = []string{"uipathcli", "autocomplete", "complete", "--command", "upathcli service-a p"}
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
	t.Run("aiappmanager", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "aiappmanager", "get-apps") })
	t.Run("aideployer", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "aideployer", "update-mlskill") })
	t.Run("aihelper", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "aihelper", "get-project-acces") })
	t.Run("aitrainer", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "aitrainer", "create-dataset") })
	t.Run("digitizer", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "digitizer", "digitize") })
	t.Run("events", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "events", "create-subscription") })
	t.Run("metering", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "metering", "track") })
	t.Run("orchestrator", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "orchestrator", "users-get-by-id") })
	t.Run("storage", func(t *testing.T) { MainParsesBuiltInDefinitions(t, "storage", "presigned-url") })
}

func MainParsesBuiltInDefinitions(t *testing.T, name string, expected string) {
	definitionDir, _ := os.Getwd()
	t.Setenv("UIPATHCLI_DEFINITIONS_PATH", filepath.Join(definitionDir, "definitions/"))

	os.Args = []string{"uipathcli", name}
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
		defer reader.Close()
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

func createFile(t *testing.T, directory string, name string) string {
	extensions := strings.TrimPrefix(filepath.Ext(name), ".")
	filename := strings.TrimSuffix(name, filepath.Ext(name)) + ".*" + extensions
	path := filepath.Join(os.TempDir(), randomDirectoryName(), directory, filename)
	err := os.MkdirAll(filepath.Dir(path), 0700)
	if err != nil {
		t.Fatal(err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(path), filename)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tempFile.Name()) })
	return tempFile.Name()
}

func randomDirectoryName() string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return hex.EncodeToString(randBytes)
}
