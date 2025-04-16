package studio

import (
	"strings"
)

type StudioProject struct {
	Name            string
	Description     string
	ProjectId       string
	TargetFramework TargetFramework
}

func (p StudioProject) NupkgIgnoreFilePattern() string {
	id := strings.ReplaceAll(p.Name, " ", ".")
	return id + ".*.nupkg"
}

func NewStudioProject(name string, description string, projectId string, targetFramework TargetFramework) *StudioProject {
	return &StudioProject{name, description, projectId, targetFramework}
}
