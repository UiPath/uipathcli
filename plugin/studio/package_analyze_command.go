package studio

import (
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
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/directories"
	"github.com/UiPath/uipathcli/utils/process"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The PackageAnalyzeCommand runs static code analyis on the project to detect common errors.
type PackageAnalyzeCommand struct {
	Exec process.ExecProcess
}

func (c PackageAnalyzeCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("package", "Package", "UiPath Studio package-related actions").
		WithOperation("analyze", "Analyze Project", "Runs static code analysis on the project to detect common errors").
		WithParameter("source", plugin.ParameterTypeString, "Path to a project.json file or a folder containing project.json file (default: .)", false).
		WithParameter("stop-on-rule-violation", plugin.ParameterTypeBoolean, "Fail when any rule is violated (default: true)", false).
		WithParameter("treat-warnings-as-errors", plugin.ParameterTypeBoolean, "Treat warnings as errors", false).
		WithParameter("governance-file", plugin.ParameterTypeString, "Pass governance policies containing the Workflow Analyzer rules (default: uipath.policy.default.json)", false)
}

func (c PackageAnalyzeCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	source, err := c.getSource(ctx)
	if err != nil {
		return err
	}

	stopOnRuleViolation := c.getBoolParameter("stop-on-rule-violation", true, ctx.Parameters)
	treatWarningsAsErrors := c.getBoolParameter("treat-warnings-as-errors", false, ctx.Parameters)
	governanceFile, err := c.getGovernanceFile(ctx, source)
	if err != nil {
		return err
	}

	exitCode, result, err := c.execute(source, stopOnRuleViolation, treatWarningsAsErrors, governanceFile, ctx.Debug, logger)
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

func (c PackageAnalyzeCommand) execute(source string, stopOnRuleViolation bool, treatWarningsAsErrors bool, governanceFile string, debug bool, logger log.Logger) (int, *packageAnalyzeResult, error) {
	jsonResultFilePath, err := c.getTemporaryJsonResultFilePath()
	if err != nil {
		return 1, nil, err
	}
	defer os.Remove(jsonResultFilePath)

	projectReader := newStudioProjectReader(source)
	project, err := projectReader.ReadMetadata()
	if err != nil {
		return 1, nil, err
	}
	supported, err := project.TargetFramework.IsSupported()
	if !supported {
		return 1, nil, err
	}

	uipcli := newUipcli(c.Exec, logger)
	err = uipcli.Initialize(project.TargetFramework)
	if err != nil {
		return 1, nil, err
	}

	if !debug {
		bar := c.newAnalyzingProgressBar(logger)
		defer close(bar)
	}
	args := c.prepareAnalyzeArguments(source, jsonResultFilePath, governanceFile)
	exitCode, stdErr, err := uipcli.ExecuteAndWait(args...)
	if err != nil {
		return exitCode, nil, err
	}

	violations, err := c.readAnalyzeResult(jsonResultFilePath)
	if err != nil {
		return 1, nil, err
	}
	errorViolationsFound := c.hasErrorViolations(violations, treatWarningsAsErrors)

	if exitCode != 0 {
		return exitCode, newErrorPackageAnalyzeResult(violations, stdErr), nil
	} else if stopOnRuleViolation && errorViolationsFound {
		return 1, newFailedPackageAnalyzeResult(violations), nil
	} else if errorViolationsFound {
		return 0, newFailedPackageAnalyzeResult(violations), nil
	}
	return 0, newSucceededPackageAnalyzeResult(violations), nil
}

func (c PackageAnalyzeCommand) prepareAnalyzeArguments(source string, jsonResultFilePath string, governanceFile string) []string {
	args := []string{"package", "analyze", source, "--resultPath", jsonResultFilePath}
	if governanceFile != "" {
		args = append(args, "--governanceFilePath", governanceFile)
	}
	return args
}

func (c PackageAnalyzeCommand) hasErrorViolations(violations []packageAnalyzeViolation, treatWarningsAsErrors bool) bool {
	for _, violation := range violations {
		if violation.Severity == "Error" {
			return true
		}
		if treatWarningsAsErrors && violation.Severity == "Warning" {
			return true
		}
	}
	return false
}

func (c PackageAnalyzeCommand) getTemporaryJsonResultFilePath() (string, error) {
	tempDirectory, err := directories.Temp()
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
			Severity:            c.convertToSeverity(entry.ErrorSeverity),
			Recommendation:      entry.Recommendation,
			DocumentationLink:   entry.DocumentationLink,
			ActivityId:          activityId,
			Item:                item,
		}
		violations = append(violations, violation)
	}
	return violations
}

func (c PackageAnalyzeCommand) convertToSeverity(errorSeverity int) string {
	switch errorSeverity {
	case 0:
		return "Off"
	case 1:
		return "Error"
	case 2:
		return "Warning"
	case 3:
		return "Info"
	default:
		return "Trace"
	}
}

func (c PackageAnalyzeCommand) newAnalyzingProgressBar(logger log.Logger) chan struct{} {
	progressBar := visualization.NewProgressBar(logger)
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

func (c PackageAnalyzeCommand) getSource(ctx plugin.ExecutionContext) (string, error) {
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

func (c PackageAnalyzeCommand) defaultGovernanceFile(source string) string {
	directory := filepath.Dir(source)
	file := filepath.Join(directory, "uipath.policy.default.json")
	fileInfo, err := os.Stat(file)
	if err != nil || fileInfo.IsDir() {
		return ""
	}
	return file
}

func (c PackageAnalyzeCommand) getGovernanceFile(context plugin.ExecutionContext, source string) (string, error) {
	governanceFileParam := c.getParameter("governance-file", "", context.Parameters)
	if governanceFileParam == "" {
		return c.defaultGovernanceFile(source), nil
	}

	file, _ := filepath.Abs(governanceFileParam)
	fileInfo, err := os.Stat(file)
	if err != nil || fileInfo.IsDir() {
		return "", fmt.Errorf("%s not found", governanceFileParam)
	}
	return file, nil
}

func (c PackageAnalyzeCommand) getParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func (c PackageAnalyzeCommand) getBoolParameter(name string, defaultValue bool, parameters []plugin.ExecutionParameter) bool {
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

func NewPackageAnalyzeCommand() *PackageAnalyzeCommand {
	return &PackageAnalyzeCommand{process.NewExecProcess()}
}
