package studio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/directories"
	"github.com/UiPath/uipathcli/utils/process"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The TestRunCommand packs a project as a test package,
// uploads it to the connected Orchestrator instances
// and runs the tests.
type TestRunCommand struct {
	Exec process.ExecProcess
}

func (c TestRunCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("test", "Test", "Tests your UiPath studio packages").
		WithOperation("run", "Run Tests", "Tests a given package").
		WithParameter("source", plugin.ParameterTypeString, "Path to a project.json file or a folder containing project.json file (default: .)", false).
		WithParameter("timeout", plugin.ParameterTypeInteger, "Time to wait in seconds for tests to finish (default: 3600)", false)
}

func (c TestRunCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	source, err := c.getSource(ctx)
	if err != nil {
		return err
	}
	timeout := time.Duration(c.getIntParameter("timeout", 3600, ctx.Parameters)) * time.Second
	result, err := c.execute(source, timeout, ctx, logger)
	if err != nil {
		return err
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("pack command failed: %v", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(json)))
}

func (c TestRunCommand) execute(source string, timeout time.Duration, ctx plugin.ExecutionContext, logger log.Logger) (*testRunResult, error) {
	projectReader := newStudioProjectReader(source)
	project, err := projectReader.ReadMetadata()
	if err != nil {
		return nil, err
	}
	supported, err := project.TargetFramework.IsSupported()
	if !supported {
		return nil, err
	}
	tmp, err := directories.Temp()
	if err != nil {
		return nil, err
	}
	destination := filepath.Join(tmp, ctx.Settings.OperationId)
	defer os.RemoveAll(destination)

	uipcli := newUipcli(c.Exec, logger)
	err = uipcli.Initialize(project.TargetFramework)
	if err != nil {
		return nil, err
	}

	exitCode, stdErr, err := c.executeUipcli(uipcli, source, destination, ctx.Debug, logger)
	if err != nil {
		return nil, err
	}
	if exitCode != 0 {
		return nil, fmt.Errorf("Error packaging tests: %v", stdErr)
	}

	nupkgPath := findLatestNupkg(destination)
	nupkgReader := newNupkgReader(nupkgPath)
	nuspec, err := nupkgReader.ReadNuspec()
	if err != nil {
		return nil, err
	}

	params := newTestRunParams(nupkgPath, nuspec.Id, nuspec.Version, timeout)
	execution, err := c.runTests(*params, ctx, logger)
	if err != nil {
		return nil, err
	}
	return newTestRunResult(*execution), nil
}

func (c TestRunCommand) runTests(params testRunParams, ctx plugin.ExecutionContext, logger log.Logger) (*TestExecution, error) {
	progressBar := visualization.NewProgressBar(logger)
	defer progressBar.Remove()
	progressBar.UpdatePercentage("uploading...", 0)

	baseUri := c.formatUri(ctx.BaseUri, ctx.Organization, ctx.Tenant)
	client := newOrchestratorClient(baseUri, ctx.Auth, ctx.Debug, ctx.Settings, logger)
	folderId, err := client.GetSharedFolderId()
	if err != nil {
		return nil, err
	}
	file := stream.NewFileStream(params.NupkgPath)
	err = client.Upload(file, progressBar)
	if err != nil {
		return nil, err
	}
	releaseId, err := client.CreateOrUpdateRelease(folderId, params.ProcessKey, params.ProcessVersion)
	if err != nil {
		return nil, err
	}
	testSetId, err := client.CreateTestSet(folderId, releaseId, params.ProcessVersion)
	if err != nil {
		return nil, err
	}
	executionId, err := client.ExecuteTestSet(folderId, testSetId)
	if err != nil {
		return nil, err
	}
	return client.WaitForTestExecutionToFinish(folderId, executionId, params.Timeout, func(execution TestExecution) {
		total := len(execution.TestCaseExecutions)
		completed := 0
		for _, testCase := range execution.TestCaseExecutions {
			if testCase.IsCompleted() {
				completed++
			}
		}
		progressBar.UpdateSteps("running...  ", completed, total)
	})
}

func (c TestRunCommand) executeUipcli(uipcli *uipcli, source string, destination string, debug bool, logger log.Logger) (int, string, error) {
	if !debug {
		bar := c.newPackagingProgressBar(logger)
		defer close(bar)
	}
	args := []string{"package", "pack", source, "--outputType", "Tests", "--autoVersion", "--output", destination}
	return uipcli.ExecuteAndWait(args...)
}

func (c TestRunCommand) newPackagingProgressBar(logger log.Logger) chan struct{} {
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

func (c TestRunCommand) getSource(ctx plugin.ExecutionContext) (string, error) {
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

func (c TestRunCommand) getIntParameter(name string, defaultValue int, parameters []plugin.ExecutionParameter) int {
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

func (c TestRunCommand) getParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func (c TestRunCommand) formatUri(baseUri url.URL, org string, tenant string) string {
	path := baseUri.Path
	if baseUri.Path == "" {
		path = "/{organization}/{tenant}/orchestrator_"
	}
	path = strings.ReplaceAll(path, "{organization}", org)
	path = strings.ReplaceAll(path, "{tenant}", tenant)
	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("%s://%s%s", baseUri.Scheme, baseUri.Host, path)
}

func NewTestRunCommand() *TestRunCommand {
	return &TestRunCommand{process.NewExecProcess()}
}
