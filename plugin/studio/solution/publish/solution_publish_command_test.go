package publish

import (
	"net/http"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
)

func TestPublishSolutionMissingOrganizationReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "publish", "--solution-id", "abc-123"}, context)

	if result.Error == nil || result.Error.Error() != "Organization is not set" {
		t.Errorf("Expected organization is not set error, but got: %v", result.Error)
	}
}

func TestPublishSolutionMissingSolutionIdReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionPublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "publish", "--organization", "my-org"}, context)

	if result.Error == nil {
		t.Errorf("Expected error for missing solution-id, but got none")
	}
}

func TestPublishSolutionSucceeds(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/Publish-Requests", http.StatusOK, `{"requestId":"req-456","status":"queued"}`).
		WithCommandPlugin(NewSolutionPublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "publish", "--organization", "my-org", "--solution-id", "abc-123"}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status Succeeded, but got: %v", result.StdOut)
	}
	if stdout["requestId"] != "req-456" {
		t.Errorf("Expected requestId req-456, but got: %v", stdout["requestId"])
	}
}

func TestPublishSolutionSendsJsonBody(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/Publish-Requests", http.StatusOK, `{"requestId":"req-456"}`).
		WithCommandPlugin(NewSolutionPublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "publish", "--organization", "my-org", "--solution-id", "abc-123"}, context)

	contentType := result.RequestHeader["content-type"]
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, but got: %v", contentType)
	}
	if !strings.Contains(result.RequestBody, "abc-123") {
		t.Errorf("Expected request body to contain solution id, but got: %v", result.RequestBody)
	}
}

func TestPublishSolutionAcceptedStatusSucceeds(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/Publish-Requests", http.StatusAccepted, `{"requestId":"req-789","status":"accepted"}`).
		WithCommandPlugin(NewSolutionPublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "publish", "--organization", "my-org", "--solution-id", "abc-123"}, context)

	if result.Error != nil {
		t.Errorf("Expected no error for 202 Accepted, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["requestId"] != "req-789" {
		t.Errorf("Expected requestId req-789, but got: %v", stdout["requestId"])
	}
}

func TestPublishSolutionBadRequestReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/Publish-Requests", http.StatusBadRequest, `{"error":"invalid solution"}`).
		WithCommandPlugin(NewSolutionPublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "publish", "--organization", "my-org", "--solution-id", "abc-123"}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "400") {
		t.Errorf("Expected error with status code 400, but got: %v", result.Error)
	}
}

func TestPublishSolutionServerErrorReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithResponse(http.StatusServiceUnavailable, `{}`).
		WithCommandPlugin(NewSolutionPublishCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "publish", "--organization", "my-org", "--solution-id", "abc-123"}, context)

	if result.Error == nil {
		t.Errorf("Expected error for server failure, but got none")
	}
}
