// Package push implements the command plugin for pushing a .uis solution
// file to UiPath Studio Web.
package push

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/api"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The SolutionPushCommand pushes a .uis file to Studio Web.
type SolutionPushCommand struct{}

func (c SolutionPushCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("solution", "UiPath Solution management", "Pack, unpack, push and pull UiPath Maestro solutions.").
		WithOperation("push", "Push Solution", "Pushes a .uis solution file to UiPath Studio Web").
		WithParameter(plugin.NewParameter("source", plugin.ParameterTypeString, "Path to .uis file").
			WithRequired(true)).
		WithParameter(plugin.NewParameter("solution-id", plugin.ParameterTypeString, "Solution ID to update (optional, for updating existing solutions)").
			WithDefaultValue(""))
}

func (c SolutionPushCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if ctx.Organization == "" {
		return errors.New("Organization is not set")
	}
	source := c.getStringParameter("source", "", ctx.Parameters)
	if source == "" {
		return errors.New("Source .uis file is required")
	}
	source, _ = filepath.Abs(source)
	solutionId := c.getStringParameter("solution-id", "", ctx.Parameters)

	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("File not found: %s", source)
	}

	params := newSolutionPushParams(source, solutionId, ctx.BaseUri, ctx.Organization, ctx.Auth, ctx.Debug, ctx.Settings)
	result, err := c.push(*params, logger)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Push command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(jsonData)))
}

func (c SolutionPushCommand) push(params solutionPushParams, logger log.Logger) (*solutionPushResult, error) {
	file := stream.NewFileStream(params.Source)
	uploadBar := visualization.NewProgressBar(logger)
	defer uploadBar.Remove()

	client := api.NewStudioClient(params.BaseUri, params.Organization, params.Auth.Token, params.Debug, params.Settings, logger)
	response, err := client.PushSolution(file, params.SolutionId, uploadBar)
	if err != nil {
		return nil, err
	}

	return newSucceededSolutionPushResult(params.Source, response.SolutionId), nil
}

func (c SolutionPushCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func NewSolutionPushCommand() *SolutionPushCommand {
	return &SolutionPushCommand{}
}
