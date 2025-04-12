package studio

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/api"
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
		WithParameter("source", plugin.ParameterTypeStringArray, "Path to one or more project.json files or folders containing project.json files (default: .)", false).
		WithParameter("timeout", plugin.ParameterTypeInteger, "Time to wait in seconds for tests to finish (default: 3600)", false)
}

func (c TestRunCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	sources, err := c.getSources(ctx)
	if err != nil {
		return err
	}
	timeout := time.Duration(c.getIntParameter("timeout", 3600, ctx.Parameters)) * time.Second

	params, err := c.prepareExecution(sources, timeout, logger)
	if err != nil {
		return err
	}
	result, err := c.executeAll(params, ctx, logger)
	if err != nil {
		return err
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("pack command failed: %v", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(json)))
}

func (c TestRunCommand) prepareExecution(sources []string, timeout time.Duration, logger log.Logger) ([]testRunParams, error) {
	tmp, err := directories.Temp()
	if err != nil {
		return nil, err
	}

	params := []testRunParams{}
	for i, source := range sources {
		projectReader := newStudioProjectReader(source)
		project, err := projectReader.ReadMetadata()
		if err != nil {
			return nil, err
		}
		supported, err := project.TargetFramework.IsSupported()
		if !supported {
			return nil, err
		}

		executionLogger := logger
		if len(sources) > 1 {
			executionLogger = NewMultiLogger(logger, "["+strconv.Itoa(i+1)+"] ")
		}
		uipcli := newUipcli(c.Exec, executionLogger)
		err = uipcli.Initialize(project.TargetFramework)
		if err != nil {
			return nil, err
		}
		destination := filepath.Join(tmp, c.randomTestRunFolderName())
		params = append(params, *newTestRunParams(i, uipcli, executionLogger, source, destination, timeout))
	}
	return params, nil
}

func (c TestRunCommand) executeAll(params []testRunParams, ctx plugin.ExecutionContext, logger log.Logger) (*testRunReport, error) {
	statusChannel := make(chan testRunStatus)
	var wg sync.WaitGroup
	for _, p := range params {
		wg.Add(1)
		go c.execute(p, ctx, p.Logger, &wg, statusChannel)
	}

	go func() {
		wg.Wait()
		close(statusChannel)
	}()

	var progressBar *visualization.ProgressBar
	if !ctx.Debug {
		progressBar = visualization.NewProgressBar(logger)
		defer progressBar.Remove()
	}
	once := sync.Once{}
	progress := c.showPackagingProgress(progressBar)
	defer once.Do(func() { close(progress) })

	status := make([]testRunStatus, len(params))
	for s := range statusChannel {
		once.Do(func() { close(progress) })
		status[s.ExecutionId] = s
		c.updateProgressBar(progressBar, status)
	}

	results := []testRunResult{}
	for _, s := range status {
		if s.Err != nil {
			return nil, s.Err
		}
		results = append(results, *s.Result)
	}
	return newTestRunReport(results), nil
}

func (c TestRunCommand) updateProgressBar(progressBar *visualization.ProgressBar, status []testRunStatus) {
	if progressBar == nil {
		return
	}
	state, totalTests, completedTests := c.calculateOverallProgress(status)
	if state == TestRunStatusUploading {
		progressBar.UpdatePercentage("uploading...", 0)
	} else if state == TestRunStatusRunning && totalTests == 0 && completedTests == 0 {
		progressBar.UpdatePercentage("running...  ", 0)
	} else if state == TestRunStatusRunning {
		progressBar.UpdateSteps("running...  ", completedTests, totalTests)
	}
}

func (c TestRunCommand) calculateOverallProgress(status []testRunStatus) (state string, totalTests int, completedTests int) {
	state = TestRunStatusPackaging
	for _, s := range status {
		totalTests += s.TotalTests
		completedTests += s.CompletedTests
		if state == TestRunStatusPackaging && s.State == TestRunStatusUploading {
			state = TestRunStatusUploading
		} else if s.State == TestRunStatusRunning {
			state = TestRunStatusRunning
		}
	}
	return state, totalTests, completedTests
}

func (c TestRunCommand) execute(params testRunParams, ctx plugin.ExecutionContext, logger log.Logger, wg *sync.WaitGroup, status chan<- testRunStatus) {
	defer wg.Done()
	defer os.RemoveAll(params.Destination)
	packParams := newPackagePackParams(
		ctx.Organization,
		ctx.Tenant,
		ctx.BaseUri,
		ctx.Auth.Token,
		params.Source,
		params.Destination,
		"",
		true,
		"Tests",
		false,
		"")
	args := c.preparePackArguments(*packParams)
	exitCode, stdErr, err := params.Uipcli.ExecuteAndWait(args...)
	if err != nil {
		status <- *newTestRunStatusError(params.ExecutionId, err)
		return
	}
	if exitCode != 0 {
		status <- *newTestRunStatusError(params.ExecutionId, fmt.Errorf("Error packaging tests: %v", stdErr))
		return
	}

	nupkgPath := findLatestNupkg(params.Destination)
	nupkgReader := newNupkgReader(nupkgPath)
	nuspec, err := nupkgReader.ReadNuspec()
	if err != nil {
		status <- *newTestRunStatusError(params.ExecutionId, err)
		return
	}

	execution, err := c.runTests(params.ExecutionId, nupkgPath, nuspec.Id, nuspec.Version, params.Timeout, ctx, logger, status)
	if err != nil {
		status <- *newTestRunStatusError(params.ExecutionId, err)
		return
	}
	result := newTestRunResult(*execution)
	status <- *newTestRunStatusDone(params.ExecutionId, result.TestCasesCount, result)
}

func (c TestRunCommand) runTests(executionId int, nupkgPath string, processKey string, processVersion string, timeout time.Duration, ctx plugin.ExecutionContext, logger log.Logger, status chan<- testRunStatus) (*api.TestExecution, error) {
	status <- *newTestRunStatusUploading(executionId)
	baseUri := c.formatUri(ctx.BaseUri, ctx.Organization, ctx.Tenant)
	client := api.NewOrchestratorClient(baseUri, ctx.Auth.Token, ctx.Debug, ctx.Settings, logger)
	folderId, err := client.GetSharedFolderId()
	if err != nil {
		return nil, err
	}
	file := stream.NewFileStream(nupkgPath)
	err = client.Upload(file, nil)
	if err != nil {
		return nil, err
	}
	releaseId, err := client.CreateOrUpdateRelease(folderId, processKey, processVersion)
	if err != nil {
		return nil, err
	}
	testSetId, err := client.CreateTestSet(folderId, releaseId, processVersion)
	if err != nil {
		return nil, err
	}
	testExecutionId, err := client.ExecuteTestSet(folderId, testSetId)
	if err != nil {
		return nil, err
	}
	return client.WaitForTestExecutionToFinish(folderId, testExecutionId, timeout, func(execution api.TestExecution) {
		total := len(execution.TestCaseExecutions)
		completed := 0
		for _, testCase := range execution.TestCaseExecutions {
			if testCase.IsCompleted() {
				completed++
			}
		}
		status <- *newTestRunStatusRunning(executionId, total, completed)
	})
}

func (c TestRunCommand) preparePackArguments(params packagePackParams) []string {
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
		args = append(args, "--libraryOrchestratorUrl", params.BaseUri.String())
		args = append(args, "--libraryOrchestratorAuthToken", params.AuthToken.Value)
		args = append(args, "--libraryOrchestratorAccountName", params.Organization)
		if params.Tenant != "" {
			args = append(args, "--libraryOrchestratorTenant", params.Tenant)
		}
	}
	return args
}

func (c TestRunCommand) showPackagingProgress(progressBar *visualization.ProgressBar) chan struct{} {
	ticker := time.NewTicker(10 * time.Millisecond)
	cancel := make(chan struct{})
	if progressBar == nil {
		return cancel
	}

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
				return
			}
		}
	}()
	return cancel
}

func (c TestRunCommand) getSources(ctx plugin.ExecutionContext) ([]string, error) {
	sources := c.getStringArrayParameter("source", []string{"."}, ctx.Parameters)
	result := []string{}
	for _, source := range sources {
		source, _ = filepath.Abs(source)
		fileInfo, err := os.Stat(source)
		if err != nil {
			return []string{}, fmt.Errorf("%s not found", defaultProjectJson)
		}
		if fileInfo.IsDir() {
			source = filepath.Join(source, defaultProjectJson)
		}
		result = append(result, source)
	}
	return result, nil
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

func (c TestRunCommand) getStringArrayParameter(name string, defaultValue []string, parameters []plugin.ExecutionParameter) []string {
	result := defaultValue
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.([]string); ok {
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

func (c TestRunCommand) randomTestRunFolderName() string {
	value, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	return "testrun-" + value.String()
}

func NewTestRunCommand() *TestRunCommand {
	return &TestRunCommand{process.NewExecProcess()}
}
