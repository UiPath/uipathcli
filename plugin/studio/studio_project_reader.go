package studio

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type studioProjectReader struct {
	Path string
}

func (r studioProjectReader) ReadMetadata() (studioProject, error) {
	data, err := r.readProjectJson()
	if err != nil {
		return studioProject{}, err
	}
	var projectJson studioProjectJson
	err = json.Unmarshal(data, &projectJson)
	if err != nil {
		return studioProject{}, fmt.Errorf("Error parsing %s file: %v", defaultProjectJson, err)
	}
	project := newStudioProject(
		projectJson.Name,
		projectJson.Description,
		projectJson.ProjectId,
		r.convertToTargetFramework(projectJson.TargetFramework))
	return *project, nil
}

func (r studioProjectReader) convertToTargetFramework(targetFramework string) TargetFramework {
	if strings.EqualFold(targetFramework, "legacy") {
		return TargetFrameworkLegacy
	}
	if strings.EqualFold(targetFramework, "windows") {
		return TargetFrameworkWindows
	}
	return TargetFrameworkCrossPlatform
}

func (r studioProjectReader) AddToIgnoredFiles(fileName string) error {
	data, err := r.readProjectJson()
	if err != nil {
		return err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("Error parsing %s file: %v", defaultProjectJson, err)
	}

	changed, err := r.addToIgnoredFiles(result.(map[string]interface{}), fileName)
	if err != nil {
		return fmt.Errorf("Error updating %s file: %v", defaultProjectJson, err)
	}
	if !changed {
		return nil
	}
	return r.updateProjectJson(result)
}

func (r studioProjectReader) readProjectJson() ([]byte, error) {
	file, err := os.Open(r.Path)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s file: %v", defaultProjectJson, err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading %s file: %v", defaultProjectJson, err)
	}
	return data, err
}

func (r studioProjectReader) updateProjectJson(result interface{}) error {
	updated, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Error updating %s file: %v", defaultProjectJson, err)
	}
	fileInfo, err := os.Stat(r.Path)
	if err != nil {
		return fmt.Errorf("Error updating %s file: %v", defaultProjectJson, err)
	}
	err = os.WriteFile(r.Path, updated, fileInfo.Mode())
	if err != nil {
		return fmt.Errorf("Error updating %s file: %v", defaultProjectJson, err)
	}
	return nil
}

func (r studioProjectReader) addToIgnoredFiles(result map[string]interface{}, fileName string) (bool, error) {
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

func (r studioProjectReader) createObjectField(result map[string]interface{}, fieldName string) (map[string]interface{}, error) {
	if _, ok := result[fieldName]; !ok {
		result[fieldName] = map[string]interface{}{}
	}
	field, ok := result[fieldName].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Unexpected type for field '%s'", fieldName)
	}
	return field, nil
}

func (r studioProjectReader) createArrayField(result map[string]interface{}, fieldName string) ([]interface{}, error) {
	if _, ok := result[fieldName]; !ok {
		result[fieldName] = []interface{}{}
	}
	field, ok := result[fieldName].([]interface{})
	if !ok {
		return nil, fmt.Errorf("Unexpected type for field '%s'", fieldName)
	}
	return field, nil
}

func (r studioProjectReader) isFileNameIgnored(ignoredFiles []interface{}, fileName string) bool {
	for _, ignoredFile := range ignoredFiles {
		if ignoredFile == fileName {
			return true
		}
	}
	return false
}

func newStudioProjectReader(path string) *studioProjectReader {
	return &studioProjectReader{path}
}
