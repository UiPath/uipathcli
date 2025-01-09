package studio

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
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

// The PackageAnalyzeCommand runs static code analyis on the project to detect common errors.
type PackageAnalyzeCommand struct {
	Exec utils.ExecProcess
}

func (c PackageAnalyzeCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("package", "Package", "UiPath Studio package-related actions").
		WithOperation("analyze", "Analyze Project", "Runs static code analysis on the project to detect common errors").
		WithParameter("source", plugin.ParameterTypeString, "Path to a project.json file or a folder containing project.json file", true).
		WithParameter("treat-warnings-as-errors", plugin.ParameterTypeBoolean, "Treat warnings as errors", false).
		WithParameter("stop-on-rule-violation", plugin.ParameterTypeBoolean, "Fail when any rule is violated", false)
}

func (c PackageAnalyzeCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	source, err := c.getSource(context)
	if err != nil {
		return err
	}
	treatWarningsAsErrors := c.getBoolParameter("treat-warnings-as-errors", context.Parameters)
	stopOnRuleViolation := c.getBoolParameter("stop-on-rule-violation", context.Parameters)
	exitCode, result, err := c.execute(source, treatWarningsAsErrors, stopOnRuleViolation, context.Debug, logger)
	if err != nil {
		return err
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("analyze command failed: %v", err)
	}
	err = writer.WriteResponse(*output.NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(json)))
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return errors.New("")
	}
	return nil
}

func (c PackageAnalyzeCommand) execute(source string, treatWarningsAsErrors bool, stopOnRuleViolation bool, debug bool, logger log.Logger) (int, *packageAnalyzeResult, error) {
	if !debug {
		bar := c.newAnalyzingProgressBar(logger)
		defer close(bar)
	}

	jsonResultFilePath, err := c.getTemporaryJsonResultFilePath()
	if err != nil {
		return 1, nil, err
	}
	defer os.Remove(jsonResultFilePath)

	args := []string{"package", "analyze", source, "--resultPath", jsonResultFilePath}
	if treatWarningsAsErrors {
		args = append(args, "--treatWarningsAsErrors")
	}
	if stopOnRuleViolation {
		args = append(args, "--stopOnRuleViolation")
	}

	projectReader := newStudioProjectReader(source)

	uipcli := newUipcli(c.Exec, logger)
	cmd, err := uipcli.Execute(projectReader.GetTargetFramework(), args...)
	if err != nil {
		return 1, nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 1, nil, fmt.Errorf("Could not run analyze command: %v", err)
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 1, nil, fmt.Errorf("Could not run analyze command: %v", err)
	}
	defer stderr.Close()
	err = cmd.Start()
	if err != nil {
		return 1, nil, fmt.Errorf("Could not run analyze command: %v", err)
	}

	stderrOutputBuilder := new(strings.Builder)
	stderrReader := io.TeeReader(stderr, stderrOutputBuilder)

	var wg sync.WaitGroup
	wg.Add(3)
	go c.readOutput(stdout, logger, &wg)
	go c.readOutput(stderrReader, logger, &wg)
	go c.wait(cmd, &wg)
	wg.Wait()

	violations, err := c.readAnalyzeResult(jsonResultFilePath)
	if err != nil {
		return 1, nil, err
	}

	exitCode := cmd.ExitCode()
	var result *packageAnalyzeResult
	if exitCode == 0 {
		result = newSucceededPackageAnalyzeResult(violations)
	} else {
		result = newFailedPackageAnalyzeResult(
			violations,
			stderrOutputBuilder.String(),
		)
	}
	return exitCode, result, nil
}

func (c PackageAnalyzeCommand) getTemporaryJsonResultFilePath() (string, error) {
	tempDirectory, err := utils.Directories{}.Temp()
	if err != nil {
		return "", err
	}
	fileName := c.randomJsonResultFileName()
	return filepath.Join(tempDirectory, fileName), nil
}

func (c PackageAnalyzeCommand) randomJsonResultFileName() string {
	value, _ := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	return "analyzeresult-" + value.String() + ".json"
}

func (c PackageAnalyzeCommand) readAnalyzeResult(path string) ([]packageAnalyzeViolation, error) {
	file, err := os.Open(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return []packageAnalyzeViolation{}, nil
	}
	if err != nil {
		return []packageAnalyzeViolation{}, fmt.Errorf("Error reading %s file: %v", filepath.Base(path), err)
	}
	defer file.Close()
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return []packageAnalyzeViolation{}, fmt.Errorf("Error reading %s file: %v", filepath.Base(path), err)
	}

	var result analyzeResultJson
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return []packageAnalyzeViolation{}, fmt.Errorf("Error parsing %s file: %v", filepath.Base(path), err)
	}
	return c.convertToViolations(result), nil
}

func (c PackageAnalyzeCommand) convertToViolations(json analyzeResultJson) []packageAnalyzeViolation {
	violations := []packageAnalyzeViolation{}
	for _, entry := range json {
		var activityId *packageAnalyzeActivityId
		if entry.ActivityId != nil {
			activityId = &packageAnalyzeActivityId{
				Id:    entry.ActivityId.Id,
				IdRef: entry.ActivityId.IdRef,
			}
		}
		var item *packageAnalyzeItem
		if entry.Item != nil {
			item = &packageAnalyzeItem{
				Name: entry.Item.Name,
				Type: entry.Item.Type,
			}
		}
		violation := packageAnalyzeViolation{
			ErrorCode:           entry.ErrorCode,
			Description:         entry.Description,
			RuleName:            entry.RuleName,
			FilePath:            entry.FilePath,
			ActivityDisplayName: entry.ActivityDisplayName,
			WorkflowDisplayName: entry.WorkflowDisplayName,
			ErrorSeverity:       entry.ErrorSeverity,
			Recommendation:      entry.Recommendation,
			DocumentationLink:   entry.DocumentationLink,
			ActivityId:          activityId,
			Item:                item,
		}
		violations = append(violations, violation)
	}
	return violations
}

func (c PackageAnalyzeCommand) wait(cmd utils.ExecCmd, wg *sync.WaitGroup) {
	defer wg.Done()
	_ = cmd.Wait()
}

func (c PackageAnalyzeCommand) newAnalyzingProgressBar(logger log.Logger) chan struct{} {
	progressBar := utils.NewProgressBar(logger)
	ticker := time.NewTicker(10 * time.Millisecond)
	cancel := make(chan struct{})
	var percent float64 = 0
	go func() {
		for {
			select {
			case <-ticker.C:
				progressBar.UpdatePercentage("analyzing...  ", percent)
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

func (c PackageAnalyzeCommand) getSource(context plugin.ExecutionContext) (string, error) {
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

func (c PackageAnalyzeCommand) readOutput(output io.Reader, logger log.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(output)
	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		logger.Log(scanner.Text())
	}
}

func (c PackageAnalyzeCommand) getParameter(name string, parameters []plugin.ExecutionParameter) string {
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

func (c PackageAnalyzeCommand) getBoolParameter(name string, parameters []plugin.ExecutionParameter) bool {
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

func NewPackageAnalyzeCommand() *PackageAnalyzeCommand {
	return &PackageAnalyzeCommand{utils.NewExecProcess()}
}
