package list

import (
	"net/http"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
)

func TestListMissingOrganizationReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewSolutionListCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "list"}, context)

	if result.Error == nil || result.Error.Error() != "Organization is not set" {
		t.Errorf("Expected organization is not set error, but got: %v", result.Error)
	}
}

func TestListReturnsSolutions(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/List", http.StatusOK, `[{"solutionId":"sol-1","name":"MySolution","status":"active"}]`).
		WithCommandPlugin(NewSolutionListCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "list", "--organization", "my-org"}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status Succeeded, but got: %v", result.StdOut)
	}
}

func TestListReturnsSolutionDetails(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/List", http.StatusOK, `[{"solutionId":"sol-1","name":"MySolution","status":"active"},{"solutionId":"sol-2","name":"OtherSolution","status":"draft"}]`).
		WithCommandPlugin(NewSolutionListCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "list", "--organization", "my-org"}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	solutions, ok := stdout["solutions"].([]interface{})
	if !ok || len(solutions) != 2 {
		t.Errorf("Expected 2 solutions, but got: %v", stdout["solutions"])
	}
}

func TestListInvalidJsonReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/List", http.StatusOK, `not-valid-json`).
		WithCommandPlugin(NewSolutionListCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "list", "--organization", "my-org"}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "invalid response body") {
		t.Errorf("Expected invalid response body error, but got: %v", result.Error)
	}
}

func TestListReturnsEmptyList(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/List", http.StatusOK, `[]`).
		WithCommandPlugin(NewSolutionListCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "list", "--organization", "my-org"}, context)

	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}
	stdout := test.ParseOutput(t, result.StdOut)
	if stdout["status"] != "Succeeded" {
		t.Errorf("Expected status Succeeded, but got: %v", result.StdOut)
	}
	solutions, ok := stdout["solutions"].([]interface{})
	if !ok || len(solutions) != 0 {
		t.Errorf("Expected empty solutions list, but got: %v", stdout["solutions"])
	}
}

func TestListBadRequestReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithUrlResponse("/my-org/studio_/backend/api/v1/ExternalSolution/List", http.StatusBadRequest, `{"error":"bad request"}`).
		WithCommandPlugin(NewSolutionListCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "list", "--organization", "my-org"}, context)

	if result.Error == nil || !strings.Contains(result.Error.Error(), "400") {
		t.Errorf("Expected error with status code 400, but got: %v", result.Error)
	}
}

func TestListServerErrorReturnsError(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithResponse(http.StatusServiceUnavailable, `{}`).
		WithCommandPlugin(NewSolutionListCommand()).
		Build()

	result := test.RunCli([]string{"studio", "solution", "list", "--organization", "my-org"}, context)

	if result.Error == nil {
		t.Errorf("Expected error for server failure, but got none")
	}
}
