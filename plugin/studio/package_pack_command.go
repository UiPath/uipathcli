package studio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/process"
	"github.com/UiPath/uipathcli/utils/visualization"
)

const defaultProjectJson = "project.json"

var OutputTypeAllowedValues = []string{"Process", "Library", "Tests", "Objects"}

// The PackagePackCommand packs a project into a single NuGet package
type PackagePackCommand struct {
	Exec process.ExecProcess
}

func (c PackagePackCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("package", "Package", "UiPath Studio package-related actions").
		WithOperation("pack", "Package Project", "Packs a project into a single package").
		WithParameter("source", plugin.ParameterTypeString, "Path to a project.json file or a folder containing project.json file (default: .)", false).
		WithParameter("destination", plugin.ParameterTypeString, "The output folder (default .)", false).
		WithParameter("package-version", plugin.ParameterTypeString, "The package version", false).
		WithParameter("auto-version", plugin.ParameterTypeBoolean, "Auto-generate package version", false).
		WithParameter("output-type", plugin.ParameterTypeString, "Force the output to a specific type."+c.formatAllowedValues(OutputTypeAllowedValues), false).
		WithParameter("split-output", plugin.ParameterTypeBoolean, "Enables the output split to runtime and design libraries", false).
		WithParameter("release-notes", plugin.ParameterTypeString, "Add release notes", false)
}

func (c PackagePackCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	source, err := c.getSource(ctx)
	if err != nil {
		return err
	}
	destination := c.getDestination(ctx)
	packageVersion := c.getParameter("package-version", "", ctx.Parameters)
	autoVersion := c.getBoolParameter("auto-version", ctx.Parameters)
	outputType := c.getParameter("output-type", "", ctx.Parameters)
	if outputType != "" && !slices.Contains(OutputTypeAllowedValues, outputType) {
		return fmt.Errorf("Invalid output type '%s', allowed values: %s", outputType, strings.Join(OutputTypeAllowedValues, ", "))
	}
	splitOutput := c.getBoolParameter("split-output", ctx.Parameters)
	releaseNotes := c.getParameter("release-notes", "", ctx.Parameters)
	params := newPackagePackParams(source, destination, packageVersion, autoVersion, outputType, splitOutput, releaseNotes)

	result, err := c.execute(*params, ctx.Debug, logger)
	if err != nil {
		return err
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("pack command failed: %v", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(json)))
}

func (c PackagePackCommand) formatAllowedValues(allowed []string) string {
	return "\n\nAllowed Values:\n- " + strings.Join(allowed, "\n- ")
}

func (c PackagePackCommand) execute(params packagePackParams, debug bool, logger log.Logger) (*packagePackResult, error) {
	projectReader := newStudioProjectReader(params.Source)
	project, err := projectReader.ReadMetadata()
	if err != nil {
		return nil, err
	}
	supported, err := project.TargetFramework.IsSupported()
	if !supported {
		return nil, err
	}
	_ = projectReader.AddToIgnoredFiles(project.NupkgIgnoreFilePattern())

	uipcli := newUipcli(c.Exec, logger)
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
		nupkgPath := findLatestNupkg(params.Destination)
		nupkgReader := newNupkgReader(nupkgPath)
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
	source := c.getParameter("source", ".", ctx.Parameters)
	source, _ = filepath.Abs(source)
	fileInfo, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("%s not found", defaultProjectJson)
	}
	if fileInfo.IsDir() {
		source = filepath.Join(source, defaultProjectJson)
	}
	return source, nil
}

func (c PackagePackCommand) getDestination(ctx plugin.ExecutionContext) string {
	destination := c.getParameter("destination", ".", ctx.Parameters)
	destination, _ = filepath.Abs(destination)
	return destination
}

func (c PackagePackCommand) getParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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
