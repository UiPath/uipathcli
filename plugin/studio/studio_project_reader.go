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

func (p studioProjectReader) GetTargetFramework() TargetFramework {
	project, _ := p.ReadMetadata()
	if strings.EqualFold(project.TargetFramework, "legacy") {
		return TargetFrameworkLegacy
	}
	if strings.EqualFold(project.TargetFramework, "windows") {
		return TargetFrameworkWindows
	}
	return TargetFrameworkCrossPlatform
}

func (p studioProjectReader) ReadMetadata() (studioProjectJson, error) {
	file, err := os.Open(p.Path)
	if err != nil {
		return studioProjectJson{}, fmt.Errorf("Error reading %s file: %v", defaultProjectJson, err)
	}
	defer file.Close()
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return studioProjectJson{}, fmt.Errorf("Error reading %s file: %v", defaultProjectJson, err)
	}

	var project studioProjectJson
	err = json.Unmarshal(byteValue, &project)
	if err != nil {
		return studioProjectJson{}, fmt.Errorf("Error parsing %s file: %v", defaultProjectJson, err)
	}
	return project, nil
}

func newStudioProjectReader(path string) *studioProjectReader {
	return &studioProjectReader{path}
}
