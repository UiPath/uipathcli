// Package publish implements the command plugin for publishing a NuGet package
// to orchestrator so that it can be executed.
package publish

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/utils/api"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The PackagePublishCommand publishes a package
type PackagePublishCommand struct {
}

func (c PackagePublishCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("package", "UiPath Studio project packaging", "Restore, analyze, package and publish your UiPath studio projects.").
		WithOperation("publish", "Publish Package", "Publishes the package to orchestrator").
		WithParameter(plugin.NewParameter("source", plugin.ParameterTypeString, "Path to package").
			WithRequired(true).
			WithDefaultValue(".")).
		WithParameter(plugin.NewParameter("folder", plugin.ParameterTypeString, "The Orchestrator Folder").
			WithDefaultValue("Shared")).
		WithParameter(plugin.NewParameter("folder-id", plugin.ParameterTypeInteger, "Folder/OrganizationUnit Id").
			WithHidden(true))
}

func (c PackagePublishCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if ctx.Organization == "" {
		return errors.New("Organization is not set")
	}
	if ctx.Tenant == "" {
		return errors.New("Tenant is not set")
	}
	source, err := c.getSource(ctx)
	if err != nil {
		return err
	}
	folder := c.getFolder(ctx.Parameters)

	nupkgReader := studio.NewNupkgReader(source)
	nuspec, err := nupkgReader.ReadNuspec()
	if err != nil {
		return err
	}
	params := newPackagePublishParams(source, folder, nuspec.Id, nuspec.Title, nuspec.Version, ctx.BaseUri, ctx.Organization, ctx.Tenant, ctx.Auth, ctx.Debug, ctx.Settings)
	result, err := c.publish(*params, logger)
	if err != nil {
		return err
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Publish command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(json)))
}

func (c PackagePublishCommand) publish(params packagePublishParams, logger log.Logger) (*packagePublishResult, error) {
	file := stream.NewFileStream(params.Source)
	uploadBar := visualization.NewProgressBar(logger)
	defer uploadBar.Remove()

	client := api.NewOrchestratorClient(params.BaseUri, params.Organization, params.Tenant, params.Auth.Token, params.Debug, params.Settings, logger)
	folderId, err := client.GetFolderId(params.Folder)
	if err != nil {
		return nil, err
	}
	feedId, err := client.GetFolderFeed(folderId)
	if err != nil {
		return nil, err
	}
	err = client.Upload(file, feedId, uploadBar)
	if errors.Is(err, api.ErrPackageAlreadyExists) {
		errorMessage := fmt.Sprintf("Package '%s' already exists", filepath.Base(params.Source))
		return newFailedPackagePublishResult(errorMessage, params.Source, params.Name, params.Description, params.Version), nil
	}
	if err != nil {
		return nil, err
	}
	releaseId, err := client.CreateOrUpdateRelease(folderId, params.Name, params.Version)
	if err != nil {
		return nil, err
	}
	return newSucceededPackagePublishResult(params.Source, params.Name, params.Description, params.Version, releaseId), nil
}

func (c PackagePublishCommand) getSource(ctx plugin.ExecutionContext) (string, error) {
	source := c.getStringParameter("source", ".", ctx.Parameters)
	source, _ = filepath.Abs(source)
	fileInfo, err := os.Stat(source)
	if err != nil {
		return "", errors.New("Package not found.")
	}
	if fileInfo.IsDir() {
		source = studio.FindLatestNupkg(source)
	}
	if source == "" {
		return "", errors.New("Could not find package to publish")
	}
	return source, nil
}

func (c PackagePublishCommand) getFolder(parameters []plugin.ExecutionParameter) string {
	folderId := c.getIntParameter("folder-id", 0, parameters)
	if folderId != 0 {
		return strconv.Itoa(folderId)
	}
	return c.getStringParameter("folder", "Shared", parameters)
}

func (c PackagePublishCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func (c PackagePublishCommand) getIntParameter(name string, defaultValue int, parameters []plugin.ExecutionParameter) int {
	result := defaultValue
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(int); ok {
				result = data
				break
			}
		}
	}
	return result
}

func NewPackagePublishCommand() *PackagePublishCommand {
	return &PackagePublishCommand{}
}
