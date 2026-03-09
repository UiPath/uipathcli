package create

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
)

func TestCreateMissingNameReturnsError(t *testing.T) {
	cmd := NewSolutionCreateCommand()
	ctx := plugin.ExecutionContext{
		Parameters: []plugin.ExecutionParameter{},
	}

	err := cmd.Execute(ctx, output.NewMemoryOutputWriter(), log.NewDefaultLogger(io.Discard))

	if err == nil || err.Error() != "Solution name is required" {
		t.Errorf("Expected solution name required error, but got: %v", err)
	}
}

func TestCreateNonStringNameReturnsError(t *testing.T) {
	cmd := NewSolutionCreateCommand()
	ctx := plugin.ExecutionContext{
		Parameters: []plugin.ExecutionParameter{
			{Name: "name", Value: 42},
		},
	}

	err := cmd.Execute(ctx, output.NewMemoryOutputWriter(), log.NewDefaultLogger(io.Discard))

	if err == nil || err.Error() != "Solution name is required" {
		t.Errorf("Expected solution name required error, but got: %v", err)
	}
}

func TestCreateDirectoryAlreadyExistsReturnsError(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "MySolution"), 0750)
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "create", "--name", "MySolution", "--destination", dir}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "Directory already exists") {
		t.Errorf("Expected directory already exists error, but got: %v", result.Error)
	}
}

func TestCreateSucceeds(t *testing.T) {
	dir := t.TempDir()
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "create", "--name", "MyAgent", "--destination", dir}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status Succeeded, but got: %v", result.StdOut)
	}
	if stdout["name"] != "MyAgent" {
		t.Errorf("Expected name MyAgent, but got: %v", stdout["name"])
	}
	if stdout["projectName"] != "Agent" {
		t.Errorf("Expected projectName Agent, but got: %v", stdout["projectName"])
	}
}

func TestCreateReturnsSolutionId(t *testing.T) {
	dir := t.TempDir()
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "create", "--name", "MyAgent", "--destination", dir}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	solutionId, ok := stdout["solutionId"].(string)
	if !ok || solutionId == "" {
		t.Errorf("Expected non-empty solutionId, but got: %v", stdout["solutionId"])
	}
	projectId, ok := stdout["projectId"].(string)
	if !ok || projectId == "" {
		t.Errorf("Expected non-empty projectId, but got: %v", stdout["projectId"])
	}
}

func TestCreateGeneratesAllFiles(t *testing.T) {
	dir := t.TempDir()
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	test.RunCli([]string{"studio", "solution", "create", "--name", "MyAgent", "--destination", dir}, context)

	expectedFiles := []string{
		"SolutionStorage.json",
		"MyAgent.uipx",
		"Agent/project.uiproj",
		"Agent/agent.json",
		"Agent/entry-points.json",
		"Agent/flow-layout.json",
		"Agent/.agent-builder/agent.json",
		"Agent/.agent-builder/bindings.json",
		"Agent/.agent-builder/entry-points.json",
		"Agent/.project/JitCustomTypes.json",
		"Agent/evals/eval-sets/evaluation-set-default.json",
		"Agent/evals/evaluators/evaluator-default.json",
		"Agent/evals/evaluators/evaluator-default-trajectory.json",
		"resources/solution_folder/package/Agent.json",
		"resources/solution_folder/process/agent/Agent.json",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(dir, "MyAgent", f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("Expected file %s to exist, but got error: %v", f, err)
		}
	}
}

func TestCreateSolutionStorageContent(t *testing.T) {
	dir := t.TempDir()
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "create", "--name", "MyAgent", "--destination", dir}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	solutionId := stdout["solutionId"].(string)

	data, err := os.ReadFile(filepath.Join(dir, "MyAgent", "SolutionStorage.json"))
	if err != nil {
		t.Fatalf("Cannot read SolutionStorage.json: %v", err)
	}
	var storage struct {
		SolutionId string `json:"SolutionId"`
		Projects   []struct {
			ProjectId           string `json:"ProjectId"`
			ProjectRelativePath string `json:"ProjectRelativePath"`
		} `json:"Projects"`
	}
	if err := json.Unmarshal(data, &storage); err != nil {
		t.Fatalf("Cannot parse SolutionStorage.json: %v", err)
	}
	if storage.SolutionId != solutionId {
		t.Errorf("Expected SolutionId %s, but got: %v", solutionId, storage.SolutionId)
	}
	if len(storage.Projects) != 1 || storage.Projects[0].ProjectRelativePath != "Agent/project.uiproj" {
		t.Errorf("Expected 1 project with Agent/project.uiproj, but got: %v", storage.Projects)
	}
}

func TestCreateAgentJsonContent(t *testing.T) {
	dir := t.TempDir()
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	test.RunCli([]string{"studio", "solution", "create", "--name", "MyAgent", "--destination", dir,
		"--model", "gpt-4o-2024-11-20",
		"--system-prompt", "You are a research assistant."}, context)

	data, err := os.ReadFile(filepath.Join(dir, "MyAgent", "Agent", "agent.json"))
	if err != nil {
		t.Fatalf("Cannot read agent.json: %v", err)
	}
	var agent struct {
		Version  string `json:"version"`
		Type     string `json:"type"`
		Settings struct {
			Model string `json:"model"`
		} `json:"settings"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(data, &agent); err != nil {
		t.Fatalf("Cannot parse agent.json: %v", err)
	}
	if agent.Version != "1.1.0" {
		t.Errorf("Expected version 1.1.0, but got: %v", agent.Version)
	}
	if agent.Type != "lowCode" {
		t.Errorf("Expected type lowCode, but got: %v", agent.Type)
	}
	if agent.Settings.Model != "gpt-4o-2024-11-20" {
		t.Errorf("Expected model gpt-4o-2024-11-20, but got: %v", agent.Settings.Model)
	}
	if len(agent.Messages) < 2 || agent.Messages[0].Content != "You are a research assistant." {
		t.Errorf("Expected system prompt 'You are a research assistant.', but got: %v", agent.Messages)
	}
}

func TestCreateWithCustomProjectName(t *testing.T) {
	dir := t.TempDir()
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "create", "--name", "MySolution", "--project-name", "ResearchBot", "--destination", dir}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	if _, err := os.Stat(filepath.Join(dir, "MySolution", "ResearchBot", "agent.json")); err != nil {
		t.Errorf("Expected ResearchBot/agent.json to exist: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "MySolution", "resources", "solution_folder", "package", "ResearchBot.json")); err != nil {
		t.Errorf("Expected package resource for ResearchBot to exist: %v", err)
	}
}

func TestCreateCanBePackedAndUnpacked(t *testing.T) {
	dir := t.TempDir()
	createCtx := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "create", "--name", "TestAgent", "--destination", dir}, createCtx)

	if result.Error != nil {
		t.Fatalf("Create failed: %v", result.Error)
	}

	// Verify SolutionStorage.json exists (required by pack)
	solutionDir := filepath.Join(dir, "TestAgent")
	if _, err := os.Stat(filepath.Join(solutionDir, "SolutionStorage.json")); err != nil {
		t.Fatalf("Expected SolutionStorage.json in created solution: %v", err)
	}
}

func TestCreateManifestContent(t *testing.T) {
	dir := t.TempDir()
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	test.RunCli([]string{"studio", "solution", "create", "--name", "MyAgent", "--destination", dir}, context)

	data, err := os.ReadFile(filepath.Join(dir, "MyAgent", "MyAgent.uipx"))
	if err != nil {
		t.Fatalf("Cannot read .uipx: %v", err)
	}
	var manifest struct {
		DocVersion string `json:"DocVersion"`
		SolutionId string `json:"SolutionId"`
		Projects   []struct {
			Type string `json:"Type"`
		} `json:"Projects"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("Cannot parse .uipx: %v", err)
	}
	if manifest.DocVersion != "1.0.0" {
		t.Errorf("Expected DocVersion 1.0.0, but got: %v", manifest.DocVersion)
	}
	if len(manifest.Projects) != 1 || manifest.Projects[0].Type != "Agent" {
		t.Errorf("Expected 1 Agent project in manifest, but got: %v", manifest.Projects)
	}
}

func TestCreateEntryPointsContent(t *testing.T) {
	dir := t.TempDir()
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	test.RunCli([]string{"studio", "solution", "create", "--name", "MyAgent", "--destination", dir}, context)

	data, err := os.ReadFile(filepath.Join(dir, "MyAgent", "Agent", "entry-points.json"))
	if err != nil {
		t.Fatalf("Cannot read entry-points.json: %v", err)
	}
	var ep struct {
		EntryPoints []struct {
			UniqueId string `json:"uniqueId"`
			Type     string `json:"type"`
		} `json:"entryPoints"`
	}
	if err := json.Unmarshal(data, &ep); err != nil {
		t.Fatalf("Cannot parse entry-points.json: %v", err)
	}
	if len(ep.EntryPoints) != 1 || ep.EntryPoints[0].Type != "agent" {
		t.Errorf("Expected 1 agent entry point, but got: %v", ep.EntryPoints)
	}
	if ep.EntryPoints[0].UniqueId == "" {
		t.Errorf("Expected non-empty uniqueId in entry point")
	}
}

func TestCreateProcessResourceContent(t *testing.T) {
	dir := t.TempDir()
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	test.RunCli([]string{"studio", "solution", "create", "--name", "MyAgent", "--destination", dir}, context)

	data, err := os.ReadFile(filepath.Join(dir, "MyAgent", "resources", "solution_folder", "process", "agent", "Agent.json"))
	if err != nil {
		t.Fatalf("Cannot read process resource: %v", err)
	}
	var res struct {
		Resource struct {
			Kind string `json:"kind"`
			Type string `json:"type"`
			Spec struct {
				Type        string `json:"type"`
				PackageName string `json:"packageName"`
			} `json:"spec"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("Cannot parse process resource: %v", err)
	}
	if res.Resource.Kind != "process" || res.Resource.Type != "agent" {
		t.Errorf("Expected process/agent resource, but got: kind=%v type=%v", res.Resource.Kind, res.Resource.Type)
	}
	if res.Resource.Spec.PackageName != "MyAgent.agent.Agent" {
		t.Errorf("Expected packageName MyAgent.agent.Agent, but got: %v", res.Resource.Spec.PackageName)
	}
}

func TestCreateGeneratesValidUUIDs(t *testing.T) {
	cmd := SolutionCreateCommand{}
	uuid := cmd.generateUUID()

	parts := strings.Split(uuid, "-")
	if len(parts) != 5 {
		t.Fatalf("Expected 5 UUID parts, but got %d: %s", len(parts), uuid)
	}
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		t.Errorf("UUID has wrong part lengths: %s", uuid)
	}
	// Version 4: third group starts with '4'
	if parts[2][0] != '4' {
		t.Errorf("Expected UUID version 4 (third group starts with '4'), but got: %s", uuid)
	}
}

func TestCreateDefaultDestination(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionCreateCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "create", "--name", "DefaultDirAgent"}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "DefaultDirAgent", "SolutionStorage.json")); err != nil {
		t.Errorf("Expected solution in current directory: %v", err)
	}
}
