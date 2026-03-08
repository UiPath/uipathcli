package list

import (
	"net/http"
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
