package pack

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
)

func TestPackMissingSolutionDirectoryReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", "/tmp/not-found-dir"}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "Solution directory not found") {
		t.Errorf("Expected solution directory not found error, but got: %v", result.Error)
	}
}

func TestPackNotADirectoryReturnsError(t *testing.T) {
	path := test.CreateTempFile(t, "test")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", path}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "Source is not a directory") {
		t.Errorf("Expected not a directory error, but got: %v", result.Error)
	}
}

func TestPackMissingSolutionStorageReturnsError(t *testing.T) {
	dir := t.TempDir()
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", dir}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "SolutionStorage.json not found") {
		t.Errorf("Expected SolutionStorage.json not found error, but got: %v", result.Error)
	}
}

func TestPackCreatesUisFile(t *testing.T) {
	dir := createSolutionDirectory(t)
	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status Succeeded, but got: %v", result.StdOut)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("Expected .uis file to exist at %s, but got error: %v", outputPath, err)
	}
}

func TestPackContainsAllFiles(t *testing.T) {
	dir := createSolutionDirectory(t)
	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Cannot open .uis file: %v", err)
	}
	defer func() { _ = reader.Close() }()

	fileNames := map[string]bool{}
	for _, f := range reader.File {
		fileNames[f.Name] = true
	}

	expectedFiles := []string{
		"SolutionStorage.json",
		"Agent/agent.json",
		"Agent/project.uiproj",
		"Agent/.agent-builder/bindings.json",
	}
	for _, expected := range expectedFiles {
		if !fileNames[expected] {
			t.Errorf("Expected .uis to contain %s, but it was not found", expected)
		}
	}
}

func TestPackExcludesGitDirectory(t *testing.T) {
	dir := createSolutionDirectory(t)
	gitDir := filepath.Join(dir, ".git")
	_ = os.MkdirAll(gitDir, 0750)
	_ = os.WriteFile(filepath.Join(gitDir, "config"), []byte("test"), 0600)

	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Cannot open .uis file: %v", err)
	}
	defer func() { _ = reader.Close() }()

	for _, f := range reader.File {
		if strings.HasPrefix(f.Name, ".git/") || f.Name == ".git" {
			t.Errorf("Expected .git to be excluded, but found: %s", f.Name)
		}
	}
}

func TestPackExcludesPycache(t *testing.T) {
	dir := createSolutionDirectory(t)
	cacheDir := filepath.Join(dir, "Agent", "__pycache__")
	_ = os.MkdirAll(cacheDir, 0750)
	_ = os.WriteFile(filepath.Join(cacheDir, "module.pyc"), []byte("test"), 0600)

	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Cannot open .uis file: %v", err)
	}
	defer func() { _ = reader.Close() }()

	for _, f := range reader.File {
		if strings.Contains(f.Name, "__pycache__") || strings.HasSuffix(f.Name, ".pyc") {
			t.Errorf("Expected __pycache__ and .pyc to be excluded, but found: %s", f.Name)
		}
	}
}

func TestPackIncludesDotDirectories(t *testing.T) {
	dir := createSolutionDirectory(t)
	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Cannot open .uis file: %v", err)
	}
	defer func() { _ = reader.Close() }()

	found := false
	for _, f := range reader.File {
		if strings.Contains(f.Name, ".agent-builder/") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected .agent-builder/ to be included in the .uis file")
	}
}

func TestPackDefaultOutputName(t *testing.T) {
	dir := createSolutionDirectory(t)

	// Work from a temp directory so default output goes there
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", dir}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	pkg, ok := stdout["package"].(string)
	if !ok || !strings.HasSuffix(pkg, ".uis") {
		t.Errorf("Expected package path to end with .uis, but got: %v", pkg)
	}
}

func TestPackReadsSolutionId(t *testing.T) {
	dir := createSolutionDirectory(t)
	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["solutionId"] != "test-solution-id" {
		t.Errorf("Expected solutionId test-solution-id, but got: %v", stdout["solutionId"])
	}
}

func TestPackReportsFileSize(t *testing.T) {
	dir := createSolutionDirectory(t)
	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	size, ok := stdout["size"].(float64)
	if !ok || size <= 0 {
		t.Errorf("Expected positive size, but got: %v", stdout["size"])
	}
}

func TestPackExcludesPycFilesOutsidePycache(t *testing.T) {
	dir := createSolutionDirectory(t)
	// Create .pyc file directly in the Agent directory (not inside __pycache__)
	_ = os.WriteFile(filepath.Join(dir, "Agent", "module.pyc"), []byte("bytecode"), 0600)

	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	reader, err := zip.OpenReader(outputPath)
	if err != nil {
		t.Fatalf("Cannot open .uis file: %v", err)
	}
	defer func() { _ = reader.Close() }()

	for _, f := range reader.File {
		if strings.HasSuffix(f.Name, ".pyc") {
			t.Errorf("Expected .pyc files to be excluded, but found: %s", f.Name)
		}
	}
}

func TestPackWithInvalidSolutionStorageJson(t *testing.T) {
	dir := t.TempDir()
	// Write invalid JSON to SolutionStorage.json
	_ = os.WriteFile(filepath.Join(dir, "SolutionStorage.json"), []byte("not valid json"), 0600)

	agentDir := filepath.Join(dir, "Agent")
	_ = os.MkdirAll(agentDir, 0750)
	_ = os.WriteFile(filepath.Join(agentDir, "agent.json"), []byte(`{"type":"lowCode"}`), 0600)

	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	if result.Error != nil {
		t.Errorf("Expected no error even with invalid JSON, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["solutionId"] != "" {
		t.Errorf("Expected empty solutionId for invalid JSON, but got: %v", stdout["solutionId"])
	}
}

func TestPackWithUnreadableFileReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}
	dir := createSolutionDirectory(t)
	// Create a file without read permissions
	unreadablePath := filepath.Join(dir, "Agent", "unreadable.txt")
	_ = os.WriteFile(unreadablePath, []byte("secret"), 0600)
	_ = os.Chmod(unreadablePath, 0000)

	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "Error opening file") {
		t.Errorf("Expected error opening unreadable file, but got: %v", result.Error)
	}
	// Verify cleanup: output file should be removed
	if _, err := os.Stat(outputPath); err == nil {
		t.Errorf("Expected output file to be cleaned up after error")
	}
}

func TestPackWithUnreadableSolutionStorageJson(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}
	dir := t.TempDir()
	solutionStoragePath := filepath.Join(dir, "SolutionStorage.json")
	_ = os.WriteFile(solutionStoragePath, []byte(`{"SolutionId":"test"}`), 0600)
	// Make unreadable after stat check passes
	_ = os.Chmod(solutionStoragePath, 0000)
	defer func() { _ = os.Chmod(solutionStoragePath, 0600) }()

	agentDir := filepath.Join(dir, "Agent")
	_ = os.MkdirAll(agentDir, 0750)
	_ = os.WriteFile(filepath.Join(agentDir, "agent.json"), []byte(`{}`), 0600)

	outputPath := filepath.Join(t.TempDir(), "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	if result.Error != nil {
		t.Errorf("Expected no error (graceful fallback), but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["solutionId"] != "" {
		t.Errorf("Expected empty solutionId when file is unreadable, but got: %v", stdout["solutionId"])
	}
}

func TestPackToInvalidDestinationReturnsError(t *testing.T) {
	dir := createSolutionDirectory(t)
	// Use a path with non-existent parent directory
	outputPath := filepath.Join(t.TempDir(), "nonexistent-parent", "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPackCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pack", "--source", dir, "--destination", outputPath}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "Cannot create output file") {
		t.Errorf("Expected cannot create output file error, but got: %v", result.Error)
	}
}

func createSolutionDirectory(t *testing.T) string {
	dir := t.TempDir()

	solutionStorage := map[string]interface{}{
		"SolutionId": "test-solution-id",
		"Projects": []map[string]interface{}{
			{"ProjectId": "test-project-id", "ProjectRelativePath": "Agent/project.uiproj"},
		},
	}
	data, _ := json.Marshal(solutionStorage)
	_ = os.WriteFile(filepath.Join(dir, "SolutionStorage.json"), data, 0600)

	agentDir := filepath.Join(dir, "Agent")
	_ = os.MkdirAll(agentDir, 0750)
	_ = os.WriteFile(filepath.Join(agentDir, "agent.json"), []byte(`{"type":"lowCode"}`), 0600)
	_ = os.WriteFile(filepath.Join(agentDir, "project.uiproj"), []byte(`{"ProjectType":"Agent"}`), 0600)

	builderDir := filepath.Join(agentDir, ".agent-builder")
	_ = os.MkdirAll(builderDir, 0750)
	_ = os.WriteFile(filepath.Join(builderDir, "bindings.json"), []byte(`{"version":"2.0","resources":[]}`), 0600)

	projectDir := filepath.Join(agentDir, ".project")
	_ = os.MkdirAll(projectDir, 0750)
	_ = os.WriteFile(filepath.Join(projectDir, "JitCustomTypes.json"), []byte(`{}`), 0600)

	return dir
}
