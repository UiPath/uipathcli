package studio

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const DefaultProjectJson = "project.json"

type StudioProjectReader struct {
	Path string
}

func (r StudioProjectReader) ReadMetadata() (StudioProject, error) {
	data, err := r.readProjectJson()
	if err != nil {
		return StudioProject{}, err
	}
	var projectJson studioProjectJson
	err = json.Unmarshal(data, &projectJson)
	if err != nil {
		return StudioProject{}, fmt.Errorf("Error parsing %s file: %w", DefaultProjectJson, err)
	}
	project := NewStudioProject(
		projectJson.Name,
		projectJson.Description,
		projectJson.ProjectId,
		r.convertToTargetFramework(projectJson.TargetFramework))
	return *project, nil
}

func (r StudioProjectReader) convertToTargetFramework(targetFramework string) TargetFramework {
	if strings.EqualFold(targetFramework, "legacy") {
		return TargetFrameworkLegacy
	}
	if strings.EqualFold(targetFramework, "windows") {
		return TargetFrameworkWindows
	}
	return TargetFrameworkCrossPlatform
}

func (r StudioProjectReader) AddToIgnoredFiles(fileName string) error {
	data, err := r.readProjectJson()
	if err != nil {
		return err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("Error parsing %s file: %w", DefaultProjectJson, err)
	}

	changed, err := r.addToIgnoredFiles(result.(map[string]interface{}), fileName)
	if err != nil {
		return fmt.Errorf("Error updating %s file: %w", DefaultProjectJson, err)
	}
	if !changed {
		return nil
	}
	return r.updateProjectJson(result)
}

func (r StudioProjectReader) readProjectJson() ([]byte, error) {
	file, err := os.Open(r.Path)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s file: %w", DefaultProjectJson, err)
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s file: %w", DefaultProjectJson, err)
	}
	return data, err
}

func (r StudioProjectReader) updateProjectJson(result interface{}) error {
	updated, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Error updating %s file: %w", DefaultProjectJson, err)
	}
	fileInfo, err := os.Stat(r.Path)
	if err != nil {
		return fmt.Errorf("Error updating %s file: %w", DefaultProjectJson, err)
	}
	err = os.WriteFile(r.Path, updated, fileInfo.Mode())
	if err != nil {
		return fmt.Errorf("Error updating %s file: %w", DefaultProjectJson, err)
	}
	return nil
}

func (r StudioProjectReader) addToIgnoredFiles(result map[string]interface{}, fileName string) (bool, error) {
	designOptions, err := r.createObjectField(result, "designOptions")
	if err != nil {
		return false, err
	}
	processOptions, err := r.createObjectField(designOptions, "processOptions")
	if err != nil {
		return false, err
	}
	ignoredFiles, err := r.createArrayField(processOptions, "ignoredFiles")
	if err != nil {
		return false, err
	}
	if r.isFileNameIgnored(ignoredFiles, fileName) {
		return false, nil
	}
	processOptions["ignoredFiles"] = append(ignoredFiles, fileName)
	return true, nil
}

func (r StudioProjectReader) createObjectField(result map[string]interface{}, fieldName string) (map[string]interface{}, error) {
	if _, ok := result[fieldName]; !ok {
		result[fieldName] = map[string]interface{}{}
	}
	field, ok := result[fieldName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Unexpected type for field '%s'", fieldName)
	}
	return field, nil
}

func (r StudioProjectReader) createArrayField(result map[string]interface{}, fieldName string) ([]interface{}, error) {
	if _, ok := result[fieldName]; !ok {
		result[fieldName] = []interface{}{}
	}
	field, ok := result[fieldName].([]interface{})
	if !ok {
		return nil, fmt.Errorf("Unexpected type for field '%s'", fieldName)
	}
	return field, nil
}

func (r StudioProjectReader) isFileNameIgnored(ignoredFiles []interface{}, fileName string) bool {
	for _, ignoredFile := range ignoredFiles {
		if ignoredFile == fileName {
			return true
		}
	}
	return false
}

func NewStudioProjectReader(path string) *StudioProjectReader {
	return &StudioProjectReader{path}
}
