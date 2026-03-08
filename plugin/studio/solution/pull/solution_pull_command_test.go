package pull

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
