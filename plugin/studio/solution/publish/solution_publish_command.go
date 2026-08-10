// Package publish implements the command plugin for publishing a solution
// in UiPath Studio Web for deployment.
package publish

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/api"
)

// The SolutionPublishCommand publishes a solution in Studio Web.
type SolutionPublishCommand struct{}

func (c SolutionPublishCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("solution", "UiPath Solution management", "Pack, unpack, push and pull UiPath Maestro solutions.").
		WithOperation("publish", "Publish Solution", "Publishes a solution in UiPath Studio Web for deployment").
		WithParameter(plugin.NewParameter("solution-id", plugin.ParameterTypeString, "Solution ID to publish").
			WithRequired(true))
}

func (c SolutionPublishCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if ctx.Organization == "" {
		return errors.New("Organization is not set")
	}
	solutionId := c.getStringParameter("solution-id", "", ctx.Parameters)
	if solutionId == "" {
		return errors.New("Solution ID is required")
	}

	client := api.NewStudioClient(ctx.BaseUri, ctx.Organization, ctx.Auth.Token, ctx.Debug, ctx.Settings, logger)
	response, err := client.PublishSolution(solutionId)
	if err != nil {
		return err
	}

	result := solutionPublishResult{
		Status:    "Succeeded",
		RequestId: response.RequestId,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Publish command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(jsonData)))
}

func (c SolutionPublishCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
	result := defaultValue
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(string); ok {
				result = data
				break
			}
		}
	}
	return result
}

func NewSolutionPublishCommand() *SolutionPublishCommand {
	return &SolutionPublishCommand{}
}
