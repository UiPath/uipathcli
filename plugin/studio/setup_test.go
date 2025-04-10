package studio

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
)

const studioDefinition = `
openapi: 3.0.1
info:
  title: UiPath Studio
  description: UiPath Studio
  version: v1
servers:
  - url: https://cloud.uipath.com/{organization}/studio_/backend
    description: The production url
    variables:
      organization:
        description: The organization name (or id)
        default: my-org
paths:
  {}
`

const nuspecContent = `
<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
  <metadata minClientVersion="3.3">
    <id>MyLibrary</id>
    <version>1.0.0</version>
    <title>My Library</title>
  </metadata>
</package>`

func writeFile(t *testing.T, data string) string {
	path := createFile(t)
	err := os.WriteFile(path, []byte(data), 0600)
	if err != nil {
		panic(fmt.Errorf("Error writing file '%s': %w", path, err))
	}
	return path
}

func createFile(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "uipath-test")
	if err != nil {
		t.Fatal(err)
	}
	defer tempFile.Close()
	t.Cleanup(func() {
		err := os.Remove(tempFile.Name())
		if err != nil {
			t.Fatal(err)
		}
	})
	return tempFile.Name()
}

func createDirectory(t *testing.T) string {
	tmp, err := os.MkdirTemp("", "uipath-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tmp) })
	return tmp
}

func createNupkgArchive(t *testing.T, nuspec string) string {
	path := createFile(t)
	writeNupkgArchive(t, path, nuspec)
	return path
}

func writeNupkgArchive(t *testing.T, path string, nuspec string) {
	_ = os.MkdirAll(filepath.Dir(path), 0700)
	archive, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)
	nuspecWriter, err := zipWriter.Create("MyProcess.nuspec")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.WriteString(nuspecWriter, nuspec)
	if err != nil {
		t.Fatal(err)
	}
	err = zipWriter.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func parseOutput(t *testing.T, output string) map[string]interface{} {
	stdout := map[string]interface{}{}
	err := json.Unmarshal([]byte(output), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize command output: %v", err)
	}
	return stdout
}

func getArgumentValue(args []string, name string) string {
	index := slices.Index(args, name)
	if index == -1 {
		return ""
	}
	return args[index+1]
}

func studioCrossPlatformProjectDirectory() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "projects", "crossplatform")
}

func studioWindowsProjectDirectory() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "projects", "windows")
}
