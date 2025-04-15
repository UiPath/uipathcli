package studio

import (
	"os"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/test"
)

func TestStudioProjectReader_FileNotFound_ReturnsError(t *testing.T) {
	studioProjectReader := newStudioProjectReader("not-found")
	_, err := studioProjectReader.ReadMetadata()
	if !strings.HasPrefix(err.Error(), "Error reading project.json file") {
		t.Errorf("Should return reading error, but got: %v", err)
	}
}

func TestStudioProjectReaderInvalidJsonReturnsError(t *testing.T) {
	path := test.CreateFileWithContent(t, "INVALID")

	studioProjectReader := newStudioProjectReader(path)
	_, err := studioProjectReader.ReadMetadata()
	if !strings.HasPrefix(err.Error(), "Error parsing project.json file") {
		t.Errorf("Should return parsing error, but got: %v", err)
	}
}

func TestStudioProjectReaderReturnsMetadata(t *testing.T) {
	path := test.CreateFileWithContent(t, `
{
  "name": "My Process",
  "projectId": "5fe987d1-7495-4dc7-bc4c-feaf08600b95",
  "description": "This is my process",
  "targetFramework": "Portable"
}
`)

	studioProjectReader := newStudioProjectReader(path)
	project, err := studioProjectReader.ReadMetadata()
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	if project.Name != "My Process" {
		t.Errorf("Should return project name 'My Process', but got: %s", project.Name)
	}
	if project.Description != "This is my process" {
		t.Errorf("Should return project description 'This is my process', but got: %s", project.Description)
	}
	if project.ProjectId != "5fe987d1-7495-4dc7-bc4c-feaf08600b95" {
		t.Errorf("Should return project id '5fe987d1-7495-4dc7-bc4c-feaf08600b95', but got: %s", project.ProjectId)
	}
	if project.TargetFramework != TargetFrameworkCrossPlatform {
		t.Errorf("Should return cross platform target framework, but got: %d", project.TargetFramework)
	}
}

func TestStudioProjectReaderReturnsTargetFrameworkWindows(t *testing.T) {
	path := test.CreateFileWithContent(t, `
{
  "targetFramework": "windows"
}
`)

	studioProjectReader := newStudioProjectReader(path)
	project, err := studioProjectReader.ReadMetadata()
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	if project.TargetFramework != TargetFrameworkWindows {
		t.Errorf("Should return windows target framework, but got: %d", project.TargetFramework)
	}
}

func TestStudioProjectReaderReturnsTargetFrameworkLegacy(t *testing.T) {
	path := test.CreateFileWithContent(t, `
{
  "targetFramework": "Legacy"
}
`)

	studioProjectReader := newStudioProjectReader(path)
	project, err := studioProjectReader.ReadMetadata()
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	if project.TargetFramework != TargetFrameworkLegacy {
		t.Errorf("Should return legacy target framework, but got: %d", project.TargetFramework)
	}
}

func TestStudioProjectReaderUnknownTargetFrameworkDefaultsToCrossPlatform(t *testing.T) {
	path := test.CreateFileWithContent(t, `
{
  "targetFramework": "Unknown"
}
`)

	studioProjectReader := newStudioProjectReader(path)
	project, err := studioProjectReader.ReadMetadata()
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	if project.TargetFramework != TargetFrameworkCrossPlatform {
		t.Errorf("Should return cross platform target framework, but got: %d", project.TargetFramework)
	}
}

func TestStudioProjectReaderAddToIgnoredFilesCreatesNewSection(t *testing.T) {
	path := test.CreateFileWithContent(t, `
{
  "name": "My Process",
  "projectId": "5fe987d1-7495-4dc7-bc4c-feaf08600b95",
  "description": "This is my process",
  "targetFramework": "Portable"
}
`)

	studioProjectReader := newStudioProjectReader(path)
	err := studioProjectReader.AddToIgnoredFiles("*.nupkg")
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	_, err = studioProjectReader.ReadMetadata()
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	data, _ := os.ReadFile(path)
	json := string(data)
	if !strings.Contains(json, `"designOptions":{"processOptions":{"ignoredFiles":["*.nupkg"]}}`) {
		t.Errorf("Should contain ignored file pattern, but got: %s", json)
	}
}

func TestStudioProjectReaderAddToIgnoredFilesAddsToExistingDesignOptions(t *testing.T) {
	path := test.CreateFileWithContent(t, `
{
  "name": "My Process",
  "projectId": "5fe987d1-7495-4dc7-bc4c-feaf08600b95",
  "description": "This is my process",
  "targetFramework": "Portable",
  "designOptions": {
    "otherOptions": {}
  }
}
`)

	studioProjectReader := newStudioProjectReader(path)
	err := studioProjectReader.AddToIgnoredFiles("*.nupkg")
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	_, err = studioProjectReader.ReadMetadata()
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	data, _ := os.ReadFile(path)
	json := string(data)
	if !strings.Contains(json, `"designOptions":{"otherOptions":{},"processOptions":{"ignoredFiles":["*.nupkg"]}}`) {
		t.Errorf("Should contain ignored file pattern, but got: %s", json)
	}
}

func TestStudioProjectReaderAddToIgnoredFilesAddsToExistingProcessOptions(t *testing.T) {
	path := test.CreateFileWithContent(t, `
{
  "name": "My Process",
  "projectId": "5fe987d1-7495-4dc7-bc4c-feaf08600b95",
  "description": "This is my process",
  "targetFramework": "Portable",
  "designOptions": {
    "processOptions": {
	  "otherOptions": {}
    }
  }
}
`)

	studioProjectReader := newStudioProjectReader(path)
	err := studioProjectReader.AddToIgnoredFiles("*.nupkg")
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	_, err = studioProjectReader.ReadMetadata()
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	data, _ := os.ReadFile(path)
	json := string(data)
	if !strings.Contains(json, `"designOptions":{"processOptions":{"ignoredFiles":["*.nupkg"],"otherOptions":{}}}`) {
		t.Errorf("Should contain ignored file pattern, but got: %s", json)
	}
}

func TestStudioProjectReaderAddToIgnoredFilesAddsToExistingIgnoredFiles(t *testing.T) {
	path := test.CreateFileWithContent(t, `
{
  "name": "My Process",
  "projectId": "5fe987d1-7495-4dc7-bc4c-feaf08600b95",
  "description": "This is my process",
  "targetFramework": "Portable",
  "designOptions": {
    "processOptions": {
	  "ignoredFiles": ["my-file"]
	}
  }
}
`)

	studioProjectReader := newStudioProjectReader(path)
	err := studioProjectReader.AddToIgnoredFiles("*.nupkg")
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	_, err = studioProjectReader.ReadMetadata()
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	data, _ := os.ReadFile(path)
	json := string(data)
	if !strings.Contains(json, `"designOptions":{"processOptions":{"ignoredFiles":["my-file","*.nupkg"]}}`) {
		t.Errorf("Should contain ignored file pattern, but got: %s", json)
	}
}

func TestStudioProjectReaderAddToIgnoredFilesFileNotFoundReturnsError(t *testing.T) {
	studioProjectReader := newStudioProjectReader("not-found")
	err := studioProjectReader.AddToIgnoredFiles("*.nupkg")
	if !strings.HasPrefix(err.Error(), "Error reading project.json file") {
		t.Errorf("Should return reading error, but got: %v", err)
	}
}

func TestStudioProjectReaderAddToIgnoredFilesInvalidJsonReturnsError(t *testing.T) {
	path := test.CreateFileWithContent(t, "INVALID")

	studioProjectReader := newStudioProjectReader(path)
	err := studioProjectReader.AddToIgnoredFiles("*.nupkg")
	if !strings.HasPrefix(err.Error(), "Error parsing project.json file") {
		t.Errorf("Should return parsing error, but got: %v", err)
	}
}

func TestStudioProjectReaderAddToIgnoredFilesInvalidIgnoredFilesReturnsError(t *testing.T) {
	path := test.CreateFileWithContent(t, `
{
  "name": "My Process",
  "projectId": "5fe987d1-7495-4dc7-bc4c-feaf08600b95",
  "description": "This is my process",
  "targetFramework": "Portable",
  "designOptions": {
    "processOptions": {
	  "ignoredFiles": {}
	}
  }
}
`)

	studioProjectReader := newStudioProjectReader(path)
	err := studioProjectReader.AddToIgnoredFiles("*.nupkg")
	if err.Error() != "Error updating project.json file: Unexpected type for field 'ignoredFiles'" {
		t.Errorf("Should return updating error, but got: %v", err)
	}
}

func TestStudioProjectReaderAddToIgnoredFilesFilePatternExistsNoOp(t *testing.T) {
	path := test.CreateFileWithContent(t, `
{
  "name": "My Process",
  "projectId": "5fe987d1-7495-4dc7-bc4c-feaf08600b95",
  "description": "This is my process",
  "targetFramework": "Portable",
  "designOptions": {
    "processOptions": {
	  "ignoredFiles": ["my-file", "*.nupkg"]
	}
  }
}
`)

	studioProjectReader := newStudioProjectReader(path)
	err := studioProjectReader.AddToIgnoredFiles("*.nupkg")
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	_, err = studioProjectReader.ReadMetadata()
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	data, _ := os.ReadFile(path)
	json := string(data)
	if !strings.Contains(json, `["my-file", "*.nupkg"]`) {
		t.Errorf("Should contain ignored file pattern, but got: %s", json)
	}
}
