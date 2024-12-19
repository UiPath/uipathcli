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
	"strings"
	"sync"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils"
)

const defaultProjectJson = "project.json"

// The PackagePackCommand packs a project into a single NuGet package
type PackagePackCommand struct {
	Exec utils.ExecProcess
}

func (c PackagePackCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("package", "Package", "UiPath Studio package-related actions").
		WithOperation("pack", "Package Project", "Packs a project into a single package").
		WithParameter("source", plugin.ParameterTypeString, "Path to a project.json file or a folder containing project.json file", true).
		WithParameter("destination", plugin.ParameterTypeString, "The output folder", true).
		WithParameter("package-version", plugin.ParameterTypeString, "The package version", false).
		WithParameter("auto-version", plugin.ParameterTypeBoolean, "Auto-generate package version", false).
		WithParameter("output-type", plugin.ParameterTypeString, "Force the output to a specific type", false).
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
	packageVersion, _ := c.getParameter("package-version", context.Parameters)
	autoVersion, _ := c.getBoolParameter("auto-version", context.Parameters)
	outputType, _ := c.getParameter("output-type", context.Parameters)
	splitOutput, _ := c.getBoolParameter("split-output", context.Parameters)
	releaseNotes, _ := c.getParameter("release-notes", context.Parameters)
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

func (c PackagePackCommand) execute(params packagePackParams, debug bool, logger log.Logger) (*packagePackResult, error) {
	if !debug {
		bar := c.newPackagingProgressBar(logger)
		defer close(bar)
	}

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

	uipcli := newUipcli(c.Exec, logger)
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

	projectJson, err := c.readProjectJson(params.Source)
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
			projectJson.Name,
			projectJson.Description,
			projectJson.ProjectId,
			version)
	} else {
		result = newFailedPackagePackResult(
			stderrOutputBuilder.String(),
			&projectJson.Name,
			&projectJson.Description,
			&projectJson.ProjectId)
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

func (c PackagePackCommand) wait(cmd utils.ExecCmd, wg *sync.WaitGroup) {
	defer wg.Done()
	_ = cmd.Wait()
}

func (c PackagePackCommand) newPackagingProgressBar(logger log.Logger) chan struct{} {
	progressBar := utils.NewProgressBar(logger)
	ticker := time.NewTicker(10 * time.Millisecond)
	cancel := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				progressBar.Tick("packaging...  ")
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
	source, _ := c.getParameter("source", context.Parameters)
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

func (c PackagePackCommand) readProjectJson(path string) (projectJson, error) {
	file, err := os.Open(path)
	if err != nil {
		return projectJson{}, fmt.Errorf("Error reading %s file: %v", defaultProjectJson, err)
	}
	defer file.Close()
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return projectJson{}, fmt.Errorf("Error reading %s file: %v", defaultProjectJson, err)
	}

	var project projectJson
	err = json.Unmarshal(byteValue, &project)
	if err != nil {
		return projectJson{}, fmt.Errorf("Error parsing %s file: %v", defaultProjectJson, err)
	}
	return project, nil
}

func (c PackagePackCommand) getDestination(context plugin.ExecutionContext) (string, error) {
	destination, _ := c.getParameter("destination", context.Parameters)
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

func (c PackagePackCommand) getParameter(name string, parameters []plugin.ExecutionParameter) (string, error) {
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(string); ok {
				return data, nil
			}
		}
	}
	return "", fmt.Errorf("Could not find '%s' parameter", name)
}

func (c PackagePackCommand) getBoolParameter(name string, parameters []plugin.ExecutionParameter) (bool, error) {
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(bool); ok {
				return data, nil
			}
		}
	}
	return false, fmt.Errorf("Could not find '%s' parameter", name)
}

func NewPackagePackCommand() *PackagePackCommand {
	return &PackagePackCommand{utils.NewExecProcess()}
}
