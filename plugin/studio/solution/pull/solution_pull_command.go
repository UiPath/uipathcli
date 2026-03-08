// Package pull implements the command plugin for pulling a solution
// from UiPath Studio Web as a .uis file.
package pull

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/api"
)

// The SolutionPullCommand pulls a solution from Studio Web.
type SolutionPullCommand struct{}

func (c SolutionPullCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("solution", "UiPath Solution management", "Pack, unpack, push and pull UiPath Maestro solutions.").
		WithOperation("pull", "Pull Solution", "Pulls a solution from UiPath Studio Web as a .uis file").
		WithParameter(plugin.NewParameter("solution-id", plugin.ParameterTypeString, "Solution ID to pull").
			WithRequired(true)).
		WithParameter(plugin.NewParameter("destination", plugin.ParameterTypeString, "Output .uis file path").
			WithDefaultValue(""))
}

func (c SolutionPullCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if ctx.Organization == "" {
		return errors.New("Organization is not set")
	}
	solutionId := c.getStringParameter("solution-id", "", ctx.Parameters)
	if solutionId == "" {
		return errors.New("Solution ID is required")
	}
	destination := c.getStringParameter("destination", "", ctx.Parameters)
	if destination == "" {
		destination = solutionId + ".uis"
	}
	destination, _ = filepath.Abs(destination)

	params := newSolutionPullParams(solutionId, destination, ctx.BaseUri, ctx.Organization, ctx.Auth, ctx.Debug, ctx.Settings)
	result, err := c.pull(*params, logger)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Pull command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(jsonData)))
}

func (c SolutionPullCommand) pull(params solutionPullParams, logger log.Logger) (*solutionPullResult, error) {
	client := api.NewStudioClient(params.BaseUri, params.Organization, params.Auth.Token, params.Debug, params.Settings, logger)
	body, err := client.PullSolution(params.SolutionId)
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()

	outFile, err := os.Create(params.Destination)
	if err != nil {
		return nil, fmt.Errorf("Cannot create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	written, err := io.Copy(outFile, body)
	if err != nil {
		_ = os.Remove(params.Destination)
		return nil, fmt.Errorf("Error writing solution file: %w", err)
	}

	return newSucceededSolutionPullResult(params.Destination, params.SolutionId, written), nil
}

func (c SolutionPullCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func NewSolutionPullCommand() *SolutionPullCommand {
	return &SolutionPullCommand{}
}
