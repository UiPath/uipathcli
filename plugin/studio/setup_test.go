package studio

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/UiPath/uipathcli/test"
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

func findViolation(violations []interface{}, errorCode string) map[string]interface{} {
	var violation map[string]interface{}
	for _, v := range violations {
		vMap := v.(map[string]interface{})
		if vMap["errorCode"] == errorCode {
			violation = vMap
		}
	}
	return violation
}

func createLargeNupkgArchive(t *testing.T, size int) string {
	path := test.CreateFile(t)
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
	_, err = io.WriteString(nuspecWriter, nuspecContent)
	if err != nil {
		t.Fatal(err)
	}

	content, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   "Content.txt",
		Method: zip.Store,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = content.Write(make([]byte, size))
	if err != nil {
		t.Fatal(err)
	}
	err = zipWriter.Close()
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func createNupkgArchive(t *testing.T, nuspec string) string {
	path := test.CreateFile(t)
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
