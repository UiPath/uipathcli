package studio

import (
	"fmt"
	"strings"
)

type studioProject struct {
	Name            string
	Description     string
	ProjectId       string
	TargetFramework TargetFramework
}

func (p studioProject) NupkgIgnoreFilePattern() string {
	id := strings.ReplaceAll(p.Name, " ", ".")
	return fmt.Sprintf("%s.*.nupkg", id)
}

func newStudioProject(name string, description string, projectId string, targetFramework TargetFramework) *studioProject {
	return &studioProject{name, description, projectId, targetFramework}
}
