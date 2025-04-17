package publish

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
		WithCategory("package", "Package", "UiPath Studio package-related actions").
		WithOperation("publish", "Publish Package", "Publishes the package to orchestrator").
		WithParameter("source", plugin.ParameterTypeString, "Path to package (default: .)", false)
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
	nupkgReader := studio.NewNupkgReader(source)
	nuspec, err := nupkgReader.ReadNuspec()
	if err != nil {
		return err
	}
	baseUri := c.formatUri(ctx.BaseUri, ctx.Organization, ctx.Tenant)
	params := newPackagePublishParams(source, nuspec.Title, nuspec.Version, baseUri, ctx.Auth, ctx.Debug, ctx.Settings)
	result, err := c.publish(*params, logger)
	if err != nil {
		return err
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Publish command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(json)))
}

func (c PackagePublishCommand) publish(params packagePublishParams, logger log.Logger) (*packagePublishResult, error) {
	file := stream.NewFileStream(params.Source)
	uploadBar := visualization.NewProgressBar(logger)
	defer uploadBar.Remove()

	client := api.NewOrchestratorClient(params.BaseUri, params.Auth.Token, params.Debug, params.Settings, logger)
	err := client.Upload(file, uploadBar)
	if errors.Is(err, api.ErrPackageAlreadyExists) {
		errorMessage := fmt.Sprintf("Package '%s' already exists", filepath.Base(params.Source))
		return newFailedPackagePublishResult(errorMessage, params.Source, params.Name, params.Version), nil
	}
	if err != nil {
		return nil, err
	}
	return newSucceededPackagePublishResult(params.Source, params.Name, params.Version), nil
}

func (c PackagePublishCommand) getSource(ctx plugin.ExecutionContext) (string, error) {
	source := c.getParameter("source", ".", ctx.Parameters)
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

func (c PackagePublishCommand) getParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func (c PackagePublishCommand) formatUri(baseUri url.URL, org string, tenant string) string {
	path := baseUri.Path
	if baseUri.Path == "" {
		path = "/{organization}/{tenant}/orchestrator_"
	}
	path = strings.ReplaceAll(path, "{organization}", org)
	path = strings.ReplaceAll(path, "{tenant}", tenant)
	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("%s://%s%s", baseUri.Scheme, baseUri.Host, path)
}

func NewPackagePublishCommand() *PackagePublishCommand {
	return &PackagePublishCommand{}
}
