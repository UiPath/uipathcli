package unpack

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
)

func TestUnpackMissingSourceReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionUnpackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "unpack", "--source", "/tmp/not-found.uis"}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "File not found") {
		t.Errorf("Expected file not found error, but got: %v", result.Error)
	}
}

func TestUnpackInvalidZipReturnsError(t *testing.T) {
	path := test.CreateTempFile(t, "not-a-zip")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionUnpackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "unpack", "--source", path}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "Cannot open .uis file") {
		t.Errorf("Expected invalid zip error, but got: %v", result.Error)
	}
}

func TestUnpackExtractsFiles(t *testing.T) {
	uisPath := createTestUisFile(t)
	destDir := filepath.Join(t.TempDir(), "output")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionUnpackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "unpack", "--source", uisPath, "--destination", destDir}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status Succeeded, but got: %v", result.StdOut)
	}
	if stdout["solutionId"] != "test-solution-id" {
		t.Errorf("Expected solutionId test-solution-id, but got: %v", stdout["solutionId"])
	}

	// Verify files were extracted
	if _, err := os.Stat(filepath.Join(destDir, "SolutionStorage.json")); err != nil {
		t.Errorf("Expected SolutionStorage.json to be extracted, but got error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destDir, "Agent", "agent.json")); err != nil {
		t.Errorf("Expected Agent/agent.json to be extracted, but got error: %v", err)
	}
}

func TestUnpackReturnsProjectCount(t *testing.T) {
	uisPath := createTestUisFile(t)
	destDir := filepath.Join(t.TempDir(), "output")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionUnpackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "unpack", "--source", uisPath, "--destination", destDir}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	projectCount, ok := stdout["projectCount"].(float64)
	if !ok || projectCount != 1 {
		t.Errorf("Expected projectCount 1, but got: %v", stdout["projectCount"])
	}
}

func TestUnpackZipSlipReturnsError(t *testing.T) {
	uisPath := filepath.Join(t.TempDir(), "malicious.uis")
	outFile, err := os.Create(uisPath)
	if err != nil {
		t.Fatalf("Cannot create test .uis file: %v", err)
	}
	w := zip.NewWriter(outFile)
	addZipFile(t, w, "../../etc/passwd", "malicious content")
	_ = w.Close()
	_ = outFile.Close()

	destDir := filepath.Join(t.TempDir(), "output")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionUnpackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "unpack", "--source", uisPath, "--destination", destDir}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "is not allowed") {
		t.Errorf("Expected zip slip error, but got: %v", result.Error)
	}
}

func TestUnpackExtractsFileContent(t *testing.T) {
	uisPath := createTestUisFile(t)
	destDir := filepath.Join(t.TempDir(), "output")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionUnpackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "unpack", "--source", uisPath, "--destination", destDir}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	data, err := os.ReadFile(filepath.Join(destDir, "Agent", "agent.json"))
	if err != nil {
		t.Fatalf("Expected Agent/agent.json to exist: %v", err)
	}
	if string(data) != `{"type":"lowCode"}` {
		t.Errorf("Expected agent.json content to be preserved, but got: %v", string(data))
	}
}

func TestUnpackDefaultDestination(t *testing.T) {
	uisPath := createTestUisFile(t)
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionUnpackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "unpack", "--source", uisPath}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	directory, ok := stdout["directory"].(string)
	if !ok || directory == "" {
		t.Errorf("Expected directory in output, but got: %v", stdout["directory"])
	}
}

func createTestUisFile(t *testing.T) string {
	uisPath := filepath.Join(t.TempDir(), "test.uis")
	outFile, err := os.Create(uisPath)
	if err != nil {
		t.Fatalf("Cannot create test .uis file: %v", err)
	}
	defer func() { _ = outFile.Close() }()

	w := zip.NewWriter(outFile)
	defer func() { _ = w.Close() }()

	addZipFile(t, w, "SolutionStorage.json", `{"SolutionId":"test-solution-id","Projects":[{"ProjectId":"p1","ProjectRelativePath":"Agent/project.uiproj"}]}`)
	addZipFile(t, w, "Agent/agent.json", `{"type":"lowCode"}`)
	addZipFile(t, w, "Agent/project.uiproj", `{"ProjectType":"Agent"}`)

	return uisPath
}

func addZipFile(t *testing.T, w *zip.Writer, name string, content string) {
	f, err := w.Create(name)
	if err != nil {
		t.Fatalf("Cannot create zip entry: %v", err)
	}
	_, err = f.Write([]byte(content))
	if err != nil {
		t.Fatalf("Cannot write zip entry: %v", err)
	}
}
