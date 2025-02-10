package studio

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
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
		WithParameter("source", plugin.ParameterTypeString, "Path to a project.json file or a folder containing project.json file", true).
		WithParameter("destination", plugin.ParameterTypeString, "The output folder", true).
		WithParameter("package-version", plugin.ParameterTypeString, "The package version", false).
		WithParameter("auto-version", plugin.ParameterTypeBoolean, "Auto-generate package version", false).
		WithParameter("output-type", plugin.ParameterTypeString, "Force the output to a specific type."+c.formatAllowedValues(OutputTypeAllowedValues), false).
		WithParameter("split-output", plugin.ParameterTypeBoolean, "Enables the output split to runtime and design libraries", false).
		WithParameter("release-notes", plugin.ParameterTypeString, "Add release notes", false)
}

func (c PackagePackCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	source, err := c.getSource(context)
	if err != nil {
		return err
	}
	destination, err := c.getDestination(context)
	if err != nil {
		return err
	}
	packageVersion := c.getParameter("package-version", context.Parameters)
	autoVersion := c.getBoolParameter("auto-version", context.Parameters)
	outputType := c.getParameter("output-type", context.Parameters)
	if outputType != "" && !slices.Contains(OutputTypeAllowedValues, outputType) {
		return fmt.Errorf("Invalid output type '%s', allowed values: %s", outputType, strings.Join(OutputTypeAllowedValues, ", "))
	}
	splitOutput := c.getBoolParameter("split-output", context.Parameters)
	releaseNotes := c.getParameter("release-notes", context.Parameters)
	params := newPackagePackParams(source, destination, packageVersion, autoVersion, outputType, splitOutput, releaseNotes)

	result, err := c.execute(*params, context.Debug, logger)
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

	projectReader := newStudioProjectReader(params.Source)

	uipcli := newUipcli(c.Exec, logger)
	err := uipcli.Initialize(projectReader.GetTargetFramework())
	if err != nil {
		return nil, err
	}

	if !debug {
		bar := c.newPackagingProgressBar(logger)
		defer close(bar)
	}
	cmd, err := uipcli.Execute(args...)
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("Could not run pack command: %v", err)
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("Could not run pack command: %v", err)
	}
	defer stderr.Close()
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("Could not run pack command: %v", err)
	}

	stderrOutputBuilder := new(strings.Builder)
	stderrReader := io.TeeReader(stderr, stderrOutputBuilder)

	var wg sync.WaitGroup
	wg.Add(3)
	go c.readOutput(stdout, logger, &wg)
	go c.readOutput(stderrReader, logger, &wg)
	go c.wait(cmd, &wg)
	wg.Wait()

	project, err := projectReader.ReadMetadata()
	if err != nil {
		return nil, err
	}

	exitCode := cmd.ExitCode()
	var result *packagePackResult
	if exitCode == 0 {
		nupkgFile := c.findNupkg(params.Destination)
		version := c.extractVersion(nupkgFile)
		result = newSucceededPackagePackResult(
			filepath.Join(params.Destination, nupkgFile),
			project.Name,
			project.Description,
			project.ProjectId,
			version)
	} else {
		result = newFailedPackagePackResult(
			stderrOutputBuilder.String(),
			&project.Name,
			&project.Description,
			&project.ProjectId)
	}
	return result, nil
}

func (c PackagePackCommand) findNupkg(destination string) string {
	newestFile := ""
	newestTime := time.Time{}

	files, _ := os.ReadDir(destination)
	for _, file := range files {
		extension := filepath.Ext(file.Name())
		if strings.EqualFold(extension, ".nupkg") {
			fileInfo, _ := file.Info()
			time := fileInfo.ModTime()
			if time.After(newestTime) {
				newestTime = time
				newestFile = file.Name()
			}
		}
	}
	return newestFile
}

func (c PackagePackCommand) extractVersion(nupkgFile string) string {
	parts := strings.Split(nupkgFile, ".")
	len := len(parts)
	if len < 4 {
		return ""
	}
	return fmt.Sprintf("%s.%s.%s", parts[len-4], parts[len-3], parts[len-2])
}

func (c PackagePackCommand) wait(cmd process.ExecCmd, wg *sync.WaitGroup) {
	defer wg.Done()
	_ = cmd.Wait()
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

func (c PackagePackCommand) getSource(context plugin.ExecutionContext) (string, error) {
	source := c.getParameter("source", context.Parameters)
	if source == "" {
		return "", errors.New("source is not set")
	}
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

func (c PackagePackCommand) getDestination(context plugin.ExecutionContext) (string, error) {
	destination := c.getParameter("destination", context.Parameters)
	if destination == "" {
		return "", errors.New("destination is not set")
	}
	destination, _ = filepath.Abs(destination)
	return destination, nil
}

func (c PackagePackCommand) readOutput(output io.Reader, logger log.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(output)
	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		logger.Log(scanner.Text())
	}
}

func (c PackagePackCommand) getParameter(name string, parameters []plugin.ExecutionParameter) string {
	result := ""
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
