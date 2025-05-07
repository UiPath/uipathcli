package testrun

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/plugin/studio"
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
		WithParameter(plugin.NewParameter("source", plugin.ParameterTypeStringArray, "Path to one or more project.json files or folders containing project.json files").
			WithRequired(true).
			WithDefaultValue(".")).
		WithParameter(plugin.NewParameter("timeout", plugin.ParameterTypeInteger, "Time to wait in seconds for tests to finish").
			WithDefaultValue(3600)).
		WithParameter(plugin.NewParameter("results-output", plugin.ParameterTypeString, "Output type for the test results report (default: uipath)").
			WithDefaultValue("uipath").
			WithAllowedValues([]interface{}{"uipath", "junit"})).
		WithParameter(plugin.NewParameter("attach-robot-logs", plugin.ParameterTypeBoolean, "Attaches Robot Logs for each testcases along with Test Report.")).
		WithParameter(plugin.NewParameter("folder", plugin.ParameterTypeString, "The Orchestrator Folder").
			WithDefaultValue("Shared")).
		WithParameter(plugin.NewParameter("folder-id", plugin.ParameterTypeInteger, "Folder/OrganizationUnit Id").
			WithHidden(true))
}

func (c TestRunCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	sources, err := c.getSources(ctx)
	if err != nil {
		return err
	}
	timeout := time.Duration(c.getIntParameter("timeout", 3600, ctx.Parameters)) * time.Second
	resultsOutput := c.getStringParameter("results-output", "uipath", ctx.Parameters)
	attachRobotLogs := c.getBoolParameter("attach-robot-logs", false, ctx.Parameters)
	folder := c.getFolder(ctx.Parameters)

	params, err := c.prepareExecution(sources, timeout, attachRobotLogs, folder, logger)
	if err != nil {
		return err
	}
	result, err := c.executeAll(params, ctx, logger)
	if err != nil {
		return err
	}
	return c.writeOutput(ctx, result, resultsOutput, writer)
}

func (c TestRunCommand) writeOutput(ctx plugin.ExecutionContext, results []testRunStatus, resultsOutput string, writer output.OutputWriter) error {
	var data []byte
	var err error
	if resultsOutput == "uipath" {
		converter := newUiPathReportConverter()
		report := converter.Convert(results)
		data, err = json.Marshal(report)
		if err != nil {
			return fmt.Errorf("run command failed: %w", err)
		}
	} else {
		client := api.NewOrchestratorClient(ctx.BaseUri, ctx.Organization, ctx.Tenant, nil, false, ctx.Settings, nil)
		converter := newJUnitReportConverter(client)
		report := converter.Convert(results)
		data, err = xml.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("run command failed: %w", err)
		}
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(data)))
}

func (c TestRunCommand) prepareExecution(sources []string, timeout time.Duration, attachRobotLogs bool, folder string, logger log.Logger) ([]testRunParams, error) {
	tmp, err := directories.Temp()
	if err != nil {
		return nil, err
	}

	params := []testRunParams{}
	for i, source := range sources {
		projectReader := studio.NewStudioProjectReader(source)
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
		uipcli := studio.NewUipcli(c.Exec, executionLogger)
		err = uipcli.Initialize(project.TargetFramework)
		if err != nil {
			return nil, err
		}
		destination := filepath.Join(tmp, c.randomTestRunFolderName())
		params = append(params, *newTestRunParams(i, uipcli, executionLogger, source, destination, timeout, attachRobotLogs, folder))
	}
	return params, nil
}

func (c TestRunCommand) executeAll(params []testRunParams, ctx plugin.ExecutionContext, logger log.Logger) ([]testRunStatus, error) {
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

	results := []testRunStatus{}
	for _, s := range status {
		if s.Err != nil {
			return nil, s.Err
		}
		results = append(results, s)
	}
	return results, nil
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
	defer func() { _ = os.RemoveAll(params.Destination) }()
	args := c.preparePackArguments(params, ctx)
	exitCode, stdErr, err := params.Uipcli.ExecuteAndWait(args...)
	if err != nil {
		status <- *newTestRunStatusError(params.ExecutionId, err)
		return
	}
	if exitCode != 0 {
		status <- *newTestRunStatusError(params.ExecutionId, fmt.Errorf("Error packaging tests: %v", stdErr))
		return
	}

	nupkgPath := studio.FindLatestNupkg(params.Destination)
	nupkgReader := studio.NewNupkgReader(nupkgPath)
	nuspec, err := nupkgReader.ReadNuspec()
	if err != nil {
		status <- *newTestRunStatusError(params.ExecutionId, err)
		return
	}

	folderId, testSet, execution, err := c.runTests(nupkgPath, nuspec.Id, nuspec.Version, params, ctx, logger, status)
	if err != nil {
		status <- *newTestRunStatusError(params.ExecutionId, err)
		return
	}
	status <- *newTestRunStatusDone(params.ExecutionId, folderId, len(execution.TestCaseExecutions), testSet, execution)
}

func (c TestRunCommand) runTests(nupkgPath string, processKey string, processVersion string, params testRunParams, ctx plugin.ExecutionContext, logger log.Logger, status chan<- testRunStatus) (int, *api.TestSet, *api.TestExecution, error) {
	status <- *newTestRunStatusUploading(params.ExecutionId)
	client := api.NewOrchestratorClient(ctx.BaseUri, ctx.Organization, ctx.Tenant, ctx.Auth.Token, ctx.Debug, ctx.Settings, logger)
	folderId, err := client.GetFolderId(params.Folder)
	if err != nil {
		return -1, nil, nil, err
	}
	feedId, err := client.GetFolderFeed(folderId)
	if err != nil {
		return -1, nil, nil, err
	}
	file := stream.NewFileStream(nupkgPath)
	err = client.Upload(file, feedId, nil)
	if err != nil {
		return -1, nil, nil, err
	}
	releaseId, err := client.CreateOrUpdateRelease(folderId, processKey, processVersion)
	if err != nil {
		return -1, nil, nil, err
	}
	testSetId, err := client.CreateTestSet(folderId, releaseId, processVersion)
	if err != nil {
		return -1, nil, nil, err
	}
	testExecutionId, err := client.ExecuteTestSet(folderId, testSetId)
	if err != nil {
		return -1, nil, nil, err
	}
	testSet, err := client.GetTestSet(folderId, testSetId)
	if err != nil {
		return -1, nil, nil, err
	}
	testExecution, err := client.WaitForTestExecutionToFinish(folderId, testExecutionId, params.Timeout, func(execution api.TestExecution) {
		total := len(execution.TestCaseExecutions)
		completed := 0
		for _, testCase := range execution.TestCaseExecutions {
			if testCase.IsCompleted() {
				completed++
			}
		}
		status <- *newTestRunStatusRunning(params.ExecutionId, folderId, total, completed)
	})

	if testExecution != nil && params.AttachRobotLogs {
		for idx, testCase := range testExecution.TestCaseExecutions {
			robotLogs, err := client.GetRobotLogs(folderId, testCase.JobKey)
			if err != nil {
				return -1, nil, nil, err
			}
			testExecution.TestCaseExecutions[idx].SetRobotLogs(robotLogs)
		}
	}
	return folderId, testSet, testExecution, err
}

func (c TestRunCommand) preparePackArguments(params testRunParams, ctx plugin.ExecutionContext) []string {
	args := []string{"package", "pack", params.Source, "--output", params.Destination, "--autoVersion", "--outputType", "Tests"}
	if ctx.Auth.Token != nil && ctx.Organization != "" {
		args = append(args, "--libraryIdentityUrl", ctx.IdentityUri.String())
		args = append(args, "--libraryOrchestratorUrl", ctx.BaseUri.String())
		args = append(args, "--libraryOrchestratorAuthToken", ctx.Auth.Token.Value)
		args = append(args, "--libraryOrchestratorAccountName", ctx.Organization)
		if ctx.Tenant != "" {
			args = append(args, "--libraryOrchestratorTenant", ctx.Tenant)
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
			return []string{}, fmt.Errorf("%s not found", studio.DefaultProjectJson)
		}
		if fileInfo.IsDir() {
			source = filepath.Join(source, studio.DefaultProjectJson)
		}
		result = append(result, source)
	}
	return result, nil
}

func (c TestRunCommand) getFolder(parameters []plugin.ExecutionParameter) string {
	folderId := c.getIntParameter("folder-id", 0, parameters)
	if folderId != 0 {
		return strconv.Itoa(folderId)
	}
	return c.getStringParameter("folder", "Shared", parameters)
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

func (c TestRunCommand) getBoolParameter(name string, defaultValue bool, parameters []plugin.ExecutionParameter) bool {
	result := defaultValue
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

func (c TestRunCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func (c TestRunCommand) randomTestRunFolderName() string {
	value, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	return "testrun-" + value.String()
}

func NewTestRunCommand() *TestRunCommand {
	return &TestRunCommand{process.NewExecProcess()}
}
