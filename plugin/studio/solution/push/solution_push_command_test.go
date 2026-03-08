package push

import (
	"net/http"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
)

func TestPushMissingOrganizationReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPushCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "push", "--source", "test.uis"}, context)

	if result.Error == nil || result.Error.Error() != "Organization is not set" {
		t.Errorf("Expected organization is not set error, but got: %v", result.Error)
	}
}

func TestPushMissingSourceReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPushCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "push", "--organization", "my-org"}, context)

	if result.Error == nil {
		t.Errorf("Expected error for missing source, but got none")
	}
}

func TestPushFileNotFoundReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPushCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "push", "--organization", "my-org", "--source", "/tmp/not-found.uis"}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "File not found") {
		t.Errorf("Expected file not found error, but got: %v", result.Error)
	}
}

func TestPushUploadsToStudioWeb(t *testing.T) {
	path := test.CreateTempFile(t, "test-solution-content")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/Push", http.StatusOK, `{"solutionId":"abc-123","status":"ok"}`).
		WithCommandPlugin(NewSolutionPushCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "push", "--organization", "my-org", "--source", path}, context)

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
}

func TestPushSendsMultipartRequest(t *testing.T) {
	path := test.CreateTempFile(t, "test-content")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/Push", http.StatusOK, `{"solutionId":"abc-123"}`).
		WithCommandPlugin(NewSolutionPushCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "push", "--organization", "my-org", "--source", path}, context)

	contentType := result.RequestHeader["content-type"]
	if !strings.HasPrefix(contentType, "multipart/form-data; boundary=") {
		t.Errorf("Expected Content-Type to be multipart/form-data, but got: %v", contentType)
	}
}

func TestPushWithSolutionIdIncludesQueryParam(t *testing.T) {
	path := test.CreateTempFile(t, "test-content")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/Push?solutionId=existing-id", http.StatusOK, `{"solutionId":"existing-id"}`).
		WithCommandPlugin(NewSolutionPushCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "push", "--organization", "my-org", "--source", path, "--solution-id", "existing-id"}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
}

func TestPushServerErrorReturnsError(t *testing.T) {
	path := test.CreateTempFile(t, "test-content")
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithResponse(http.StatusServiceUnavailable, `{}`).
		WithCommandPlugin(NewSolutionPushCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "push", "--organization", "my-org", "--source", path}, context)

	if result.Error == nil {
		t.Errorf("Expected error for server failure, but got none")
	}
}
