package test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var projectJsonTemplate string
var mainXaml string
var defaultGovernanceFile string

func init() {
	projectJsonTemplate = readTestData("project", "project.json")
	mainXaml = readTestData("project", "main.xaml")
	defaultGovernanceFile = readTestData("project", "uipath.policy.default.json")
}

func readTestData(directory string, name string) string {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "testdata", directory, name)
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(data)
}

type ProjectBuilder struct {
	t                     *testing.T
	projectName           string
	projectId             string
	targetFramework       string
	defaultGovernanceFile string
}

func (b *ProjectBuilder) WithProjectName(name string) *ProjectBuilder {
	b.projectName = name
	return b
}

func (b *ProjectBuilder) WithDefaultGovernanceFile() *ProjectBuilder {
	b.defaultGovernanceFile = defaultGovernanceFile
	return b
}

func (b *ProjectBuilder) writeFileContent(directory string, fileName string, content string) {
	err := os.WriteFile(filepath.Join(directory, fileName), []byte(content), 0600)
	if err != nil {
		b.t.Fatal(err)
	}
}

func (b *ProjectBuilder) Build() string {
	directory := b.t.TempDir()

	projectJson := b.buildProjectJson()
	b.writeFileContent(directory, "project.json", projectJson)
	b.writeFileContent(directory, "Main.xaml", mainXaml)
	if b.defaultGovernanceFile != "" {
		b.writeFileContent(directory, "uipath.policy.default.json", b.defaultGovernanceFile)
	}

	return directory
}

func (b *ProjectBuilder) buildProjectJson() string {
	projectJson := b.formatTemplate(projectJsonTemplate, map[string]string{
		"PROJECT_ID":       b.projectId,
		"PROJECT_NAME":     b.projectName,
		"TARGET_FRAMEWORK": b.targetFramework,
	})
	return projectJson
}

func (b *ProjectBuilder) formatTemplate(template string, values map[string]string) string {
	result := template
	for key, value := range values {
		result = strings.ReplaceAll(result, "{{"+key+"}}", value)
	}
	return result
}

func NewCrossPlatformProject(t *testing.T) *ProjectBuilder {
	return &ProjectBuilder{
		t,
		"MyProcess",
		"9011ee47-8dd4-4726-8850-299bd6ef057c",
		"Portable",
		"",
	}
}

func NewWindowsProject(t *testing.T) *ProjectBuilder {
	return &ProjectBuilder{
		t,
		"MyWindowsProcess",
		"94c4c9c1-68c3-45d4-be14-d4427f17eddd",
		"Windows",
		"",
	}
}
