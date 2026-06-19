// Package list implements the command plugin for listing solutions
// from UiPath Studio Web.
package list

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

// The SolutionListCommand lists solutions from Studio Web.
type SolutionListCommand struct{}

func (c SolutionListCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("solution", "UiPath Solution management", "Pack, unpack, push and pull UiPath Maestro solutions.").
		WithOperation("list", "List Solutions", "Lists solutions from UiPath Studio Web")
}

func (c SolutionListCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if ctx.Organization == "" {
		return errors.New("Organization is not set")
	}

	client := api.NewStudioClient(ctx.BaseUri, ctx.Organization, ctx.Auth.Token, ctx.Debug, ctx.Settings, logger)
	solutions, err := client.ListSolutions()
	if err != nil {
		return err
	}

	result := solutionListResult{
		Status:    "Succeeded",
		Solutions: solutions,
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("List command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(jsonData)))
}

func NewSolutionListCommand() *SolutionListCommand {
	return &SolutionListCommand{}
}
