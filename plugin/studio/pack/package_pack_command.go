package pack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/utils/process"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The PackagePackCommand packs a project into a single NuGet package
type PackagePackCommand struct {
	Exec process.ExecProcess
}

func (c PackagePackCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("package", "Package", "UiPath Studio package-related actions").
		WithOperation("pack", "Package Project", "Packs a project into a single package").
		WithParameter(plugin.NewParameter("source", plugin.ParameterTypeString, "Path to a project.json file or a folder containing project.json file").
			WithRequired(true).
			WithDefaultValue(".")).
		WithParameter(plugin.NewParameter("destination", plugin.ParameterTypeString, "The output folder").
			WithRequired(true).
			WithDefaultValue(".")).
		WithParameter(plugin.NewParameter("package-version", plugin.ParameterTypeString, "The package version")).
		WithParameter(plugin.NewParameter("auto-version", plugin.ParameterTypeBoolean, "Auto-generate package version")).
		WithParameter(plugin.NewParameter("output-type", plugin.ParameterTypeString, "Force the output to a specific type.").
			WithAllowedValues([]interface{}{"Process", "Library", "Tests", "Objects"})).
		WithParameter(plugin.NewParameter("split-output", plugin.ParameterTypeBoolean, "Enables the output split to runtime and design libraries")).
		WithParameter(plugin.NewParameter("release-notes", plugin.ParameterTypeString, "Add release notes"))
}

func (c PackagePackCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	source, err := c.getSource(ctx)
	if err != nil {
		return err
	}
	destination := c.getDestination(ctx)
	packageVersion := c.getStringParameter("package-version", "", ctx.Parameters)
	autoVersion := c.getBoolParameter("auto-version", ctx.Parameters)
	outputType := c.getStringParameter("output-type", "", ctx.Parameters)
	splitOutput := c.getBoolParameter("split-output", ctx.Parameters)
	releaseNotes := c.getStringParameter("release-notes", "", ctx.Parameters)
	params := newPackagePackParams(
		ctx.Organization,
		ctx.Tenant,
		ctx.BaseUri,
		ctx.Auth.Token,
		ctx.IdentityUri,
		source,
		destination,
		packageVersion,
		autoVersion,
		outputType,
		splitOutput,
		releaseNotes)

	result, err := c.execute(*params, ctx.Debug, logger)
	if err != nil {
		return err
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("pack command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(json)))
}

func (c PackagePackCommand) execute(params packagePackParams, debug bool, logger log.Logger) (*packagePackResult, error) {
	projectReader := studio.NewStudioProjectReader(params.Source)
	project, err := projectReader.ReadMetadata()
	if err != nil {
		return nil, err
	}
	supported, err := project.TargetFramework.IsSupported()
	if !supported {
		return nil, err
	}
	_ = projectReader.AddToIgnoredFiles(project.NupkgIgnoreFilePattern())

	uipcli := studio.NewUipcli(c.Exec, logger)
	err = uipcli.Initialize(project.TargetFramework)
	if err != nil {
		return nil, err
	}

	if !debug {
		bar := c.newPackagingProgressBar(logger)
		defer close(bar)
	}
	args := c.preparePackArguments(params)
	exitCode, stdErr, err := uipcli.ExecuteAndWait(args...)
	if err != nil {
		return nil, err
	}

	var result *packagePackResult
	if exitCode == 0 {
		nupkgPath := studio.FindLatestNupkg(params.Destination)
		nupkgReader := studio.NewNupkgReader(nupkgPath)
		nuspec, err := nupkgReader.ReadNuspec()
		if err != nil {
			return nil, err
		}
		result = newSucceededPackagePackResult(
			nupkgPath,
			project.Name,
			project.Description,
			project.ProjectId,
			nuspec.Version)
	} else {
		result = newFailedPackagePackResult(
			stdErr,
			&project.Name,
			&project.Description,
			&project.ProjectId)
	}
	return result, nil
}

func (c PackagePackCommand) preparePackArguments(params packagePackParams) []string {
	args := []string{"package", "pack", params.Source, "--output", params.Destination}
	if params.PackageVersion != "" {
		args = append(args, "--version", params.PackageVersion)
	}
	if params.AutoVersion {
		args = append(args, "--autoVersion")
	}
	if params.OutputType != "" {
		args = append(args, "--outputType", params.OutputType)
	}
	if params.SplitOutput {
		args = append(args, "--splitOutput")
	}
	if params.ReleaseNotes != "" {
		args = append(args, "--releaseNotes", params.ReleaseNotes)
	}
	if params.AuthToken != nil && params.Organization != "" {
		args = append(args, "--libraryIdentityUrl", params.IdentityUri.String())
		args = append(args, "--libraryOrchestratorUrl", params.BaseUri.String())
		args = append(args, "--libraryOrchestratorAuthToken", params.AuthToken.Value)
		args = append(args, "--libraryOrchestratorAccountName", params.Organization)
		if params.Tenant != "" {
			args = append(args, "--libraryOrchestratorTenant", params.Tenant)
		}
	}
	return args
}

func (c PackagePackCommand) newPackagingProgressBar(logger log.Logger) chan struct{} {
	progressBar := visualization.NewProgressBar(logger)
	ticker := time.NewTicker(10 * time.Millisecond)
	cancel := make(chan struct{})
	var percent float64 = 0
	go func() {
		for {
			select {
			case <-ticker.C:
				progressBar.UpdatePercentage("packaging...  ", percent)
				percent = percent + 1
				if percent > 100 {
					percent = 0
				}
			case <-cancel:
				ticker.Stop()
				progressBar.Remove()
				return
			}
		}
	}()
	return cancel
}

func (c PackagePackCommand) getSource(ctx plugin.ExecutionContext) (string, error) {
	source := c.getStringParameter("source", ".", ctx.Parameters)
	source, _ = filepath.Abs(source)
	fileInfo, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("%s not found", studio.DefaultProjectJson)
	}
	if fileInfo.IsDir() {
		source = filepath.Join(source, studio.DefaultProjectJson)
	}
	return source, nil
}

func (c PackagePackCommand) getDestination(ctx plugin.ExecutionContext) string {
	destination := c.getStringParameter("destination", ".", ctx.Parameters)
	destination, _ = filepath.Abs(destination)
	return destination
}

func (c PackagePackCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func (c PackagePackCommand) getBoolParameter(name string, parameters []plugin.ExecutionParameter) bool {
	result := false
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(bool); ok {
				result = data
				break
			}
		}
	}
	return result
}

func NewPackagePackCommand() *PackagePackCommand {
	return &PackagePackCommand{process.NewExecProcess()}
}
