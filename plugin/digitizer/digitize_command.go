// Package digitzer implements a plugin for digitizing files
// using Document Understanding.
package digitzer

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/api"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The DigitizeCommand is a convenient wrapper over the async digitizer API
// to make it seem like it is a single sync call.
type DigitizeCommand struct{}

func (c DigitizeCommand) Command() plugin.Command {
	return *plugin.NewCommand("du").
		WithCategory("digitization", "Document Digitization", "Digitizes a document, extracting its Document Object Model (DOM) and text.").
		WithOperation("digitize", "Digitize file", "Digitize the given file").
		WithParameter(plugin.NewParameter("project-id", plugin.ParameterTypeString, "The project id")).
		WithParameter(plugin.NewParameter("file", plugin.ParameterTypeBinary, "The file to digitize").
			WithRequired(true)).
		WithParameter(plugin.NewParameter("content-type", plugin.ParameterTypeString, "The content type"))
}

func (c DigitizeCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if ctx.Organization == "" {
		return errors.New("Organization is not set")
	}
	if ctx.Tenant == "" {
		return errors.New("Tenant is not set")
	}
	documentId, err := c.startDigitization(ctx, logger)
	if err != nil {
		return err
	}

	for i := 1; i <= 60; i++ {
		finished, err := c.waitForDigitization(documentId, ctx, writer, logger)
		if err != nil {
			return err
		}
		if finished {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("Digitization with documentId '%s' did not finish in time", documentId)
}

func (c DigitizeCommand) startDigitization(ctx plugin.ExecutionContext, logger log.Logger) (string, error) {
	uploadBar := visualization.NewProgressBar(logger)
	defer uploadBar.Remove()

	projectId := c.getProjectId(ctx.Parameters)
	file := ctx.Input
	if file == nil {
		file = c.getFileParameter(ctx.Parameters)
	}
	contentType := c.getStringParameter("content-type", "application/octet-stream", ctx.Parameters)

	client := api.NewDuClient(ctx.BaseUri, ctx.Organization, ctx.Tenant, ctx.Auth.Token, ctx.Debug, ctx.Settings, logger)
	return client.StartDigitization(projectId, file, contentType, uploadBar)
}

func (c DigitizeCommand) waitForDigitization(documentId string, ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) (bool, error) {
	projectId := c.getProjectId(ctx.Parameters)
	client := api.NewDuClient(ctx.BaseUri, ctx.Organization, ctx.Tenant, ctx.Auth.Token, ctx.Debug, ctx.Settings, logger)
	result, err := client.GetDigitizationResult(projectId, documentId)
	if err != nil {
		return true, err
	}
	if result == "" {
		return false, nil
	}

	err = writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, strings.NewReader(result)))
	return true, err
}

func (c DigitizeCommand) getProjectId(parameters []plugin.ExecutionParameter) string {
	return c.getStringParameter("project-id", "00000000-0000-0000-0000-000000000000", parameters)
}

func (c DigitizeCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func (c DigitizeCommand) getFileParameter(parameters []plugin.ExecutionParameter) stream.Stream {
	var result stream.Stream
	for _, p := range parameters {
		if p.Name == "file" {
			if stream, ok := p.Value.(stream.Stream); ok {
				result = stream
				break
			}
		}
	}
	return result
}

func NewDigitizeCommand() *DigitizeCommand {
	return &DigitizeCommand{}
}
