package pull

import (
	"io"
	"net/http"
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

func TestPullMissingOrganizationReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPullCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pull", "--solution-id", "abc-123"}, context)

	if result.Error == nil || result.Error.Error() != "Organization is not set" {
		t.Errorf("Expected organization is not set error, but got: %v", result.Error)
	}
}

func TestPullMissingSolutionIdReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPullCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pull", "--organization", "my-org"}, context)

	if result.Error == nil {
		t.Errorf("Expected error for missing solution-id, but got none")
	}
}

func TestPullDownloadsSolution(t *testing.T) {
	destPath := filepath.Join(t.TempDir(), "downloaded.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/Pull?solutionId=abc-123", http.StatusOK, "fake-uis-content").
		WithCommandPlugin(NewSolutionPullCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pull", "--organization", "my-org", "--solution-id", "abc-123", "--destination", destPath}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status Succeeded, but got: %v", result.StdOut)
	}
	if stdout["solutionId"] != "abc-123" {
		t.Errorf("Expected solutionId abc-123, but got: %v", stdout["solutionId"])
	}

	data, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Expected file to be created at %s, but got error: %v", destPath, err)
	}
	if string(data) != "fake-uis-content" {
		t.Errorf("Expected file content 'fake-uis-content', but got: %v", string(data))
	}
}

func TestPullDefaultDestination(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/Pull?solutionId=abc-123", http.StatusOK, "content").
		WithCommandPlugin(NewSolutionPullCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pull", "--organization", "my-org", "--solution-id", "abc-123"}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	filePath, ok := stdout["file"].(string)
	if !ok || !strings.HasSuffix(filePath, "abc-123.uis") {
		t.Errorf("Expected file to end with abc-123.uis, but got: %v", filePath)
	}
}

func TestPullReportsFileSize(t *testing.T) {
	destPath := filepath.Join(t.TempDir(), "downloaded.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/Pull?solutionId=abc-123", http.StatusOK, "fake-uis-content").
		WithCommandPlugin(NewSolutionPullCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pull", "--organization", "my-org", "--solution-id", "abc-123", "--destination", destPath}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	size, ok := stdout["size"].(float64)
	if !ok || size <= 0 {
		t.Errorf("Expected positive size, but got: %v", stdout["size"])
	}
}

func TestPullToInvalidDestinationReturnsError(t *testing.T) {
	destPath := filepath.Join(t.TempDir(), "nonexistent-parent", "test.uis")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/Pull?solutionId=abc-123", http.StatusOK, "content").
		WithCommandPlugin(NewSolutionPullCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pull", "--organization", "my-org", "--solution-id", "abc-123", "--destination", destPath}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "Cannot create output file") {
		t.Errorf("Expected cannot create output file error, but got: %v", result.Error)
	}
}

func TestPullNotFoundReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/Pull?solutionId=not-found", http.StatusNotFound, `{"error":"Solution not found"}`).
		WithCommandPlugin(NewSolutionPullCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pull", "--organization", "my-org", "--solution-id", "not-found"}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "404") {
		t.Errorf("Expected error with status code 404, but got: %v", result.Error)
	}
}

func TestPullServerErrorReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithResponse(http.StatusServiceUnavailable, `{}`).
		WithCommandPlugin(NewSolutionPullCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "pull", "--organization", "my-org", "--solution-id", "abc-123"}, context)

	if result.Error == nil {
		t.Errorf("Expected error for server failure, but got none")
	}
}

func TestPullNonStringSolutionIdReturnsError(t *testing.T) {
	cmd := NewSolutionPullCommand()
	ctx := plugin.ExecutionContext{
		Organization: "my-org",
		Parameters: []plugin.ExecutionParameter{
			{Name: "solution-id", Value: 42},
		},
	}

	err := cmd.Execute(ctx, output.NewMemoryOutputWriter(), log.NewDefaultLogger(io.Discard))

	if err == nil || err.Error() != "Solution ID is required" {
		t.Errorf("Expected solution ID required error, but got: %v", err)
	}
}
