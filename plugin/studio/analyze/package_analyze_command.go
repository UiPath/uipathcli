// Package analyze implements the command plugin for analyzing
// UiPath Studio projects.
package analyze

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/plugin/studio"
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
		WithCategory("package", "UiPath Studio project packaging", "Restore, analyze, package and publish your UiPath studio projects.").
		WithOperation("analyze", "Analyze Project", "Runs static code analysis on the project to detect common errors").
		WithParameter(plugin.NewParameter("source", plugin.ParameterTypeString, "Path to a project.json file or a folder containing project.json file").
			WithRequired(true).
			WithDefaultValue(".")).
		WithParameter(plugin.NewParameter("stop-on-rule-violation", plugin.ParameterTypeBoolean, "Fail when any rule is violated").
			WithDefaultValue(true)).
		WithParameter(plugin.NewParameter("treat-warnings-as-errors", plugin.ParameterTypeBoolean, "Treat warnings as errors")).
		WithParameter(plugin.NewParameter("governance-file", plugin.ParameterTypeString, "Pass governance policies containing the Workflow Analyzer rules").
			WithDefaultValue("uipath.policy.default.json"))
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

	params := newPackageAnalyzeParams(
		ctx.Organization,
		ctx.Tenant,
		ctx.BaseUri,
		ctx.Auth.Token,
		ctx.IdentityUri,
		source,
		stopOnRuleViolation,
		treatWarningsAsErrors,
		governanceFile,
	)
	exitCode, result, err := c.execute(*params, ctx.Debug, logger)
	if err != nil {
		return err
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("analyze command failed: %w", err)
	}
	err = writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(json)))
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return errors.New("")
	}
	return nil
}

func (c PackageAnalyzeCommand) execute(params packageAnalyzeParams, debug bool, logger log.Logger) (int, *packageAnalyzeResult, error) {
	jsonResultFilePath, err := c.getTemporaryJsonResultFilePath()
	if err != nil {
		return 1, nil, err
	}
	defer func() { _ = os.Remove(jsonResultFilePath) }()

	projectReader := studio.NewStudioProjectReader(params.Source)
	project, err := projectReader.ReadMetadata()
	if err != nil {
		return 1, nil, err
	}
	supported, err := project.TargetFramework.IsSupported()
	if !supported {
		return 1, nil, err
	}

	uipcli := studio.NewUipcli(c.Exec, logger)
	err = uipcli.Initialize(project.TargetFramework)
	if err != nil {
		return 1, nil, err
	}

	if !debug {
		bar := c.newAnalyzingProgressBar(logger)
		defer close(bar)
	}
	args := c.prepareAnalyzeArguments(params, jsonResultFilePath)
	exitCode, stdErr, err := uipcli.ExecuteAndWait(args...)
	if err != nil {
		return exitCode, nil, err
	}

	violations, err := c.readAnalyzeResult(jsonResultFilePath)
	if err != nil {
		return 1, nil, err
	}
	errorViolationsFound := c.hasErrorViolations(violations, params.TreatWarningsAsErrors)

	if exitCode != 0 {
		return exitCode, newErrorPackageAnalyzeResult(violations, stdErr), nil
	} else if params.StopOnRuleViolation && errorViolationsFound {
		return 1, newFailedPackageAnalyzeResult(violations), nil
	} else if errorViolationsFound {
		return 0, newFailedPackageAnalyzeResult(violations), nil
	}
	return 0, newSucceededPackageAnalyzeResult(violations), nil
}

func (c PackageAnalyzeCommand) prepareAnalyzeArguments(params packageAnalyzeParams, jsonResultFilePath string) []string {
	args := []string{"package", "analyze", params.Source, "--resultPath", jsonResultFilePath}
	if params.GovernanceFile != "" {
		args = append(args, "--governanceFilePath", params.GovernanceFile)
	}
	if params.AuthToken != nil && params.Organization != "" {
		args = append(args, "--identityUrl", params.IdentityUri.String())
		args = append(args, "--orchestratorUrl", params.BaseUri.String())
		args = append(args, "--orchestratorAuthToken", params.AuthToken.Value)
		args = append(args, "--orchestratorAccountName", params.Organization)
		if params.Tenant != "" {
			args = append(args, "--orchestratorTenant", params.Tenant)
		}
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
		return []packageAnalyzeViolation{}, fmt.Errorf("Error reading %s file: %w", filepath.Base(path), err)
	}
	defer func() { _ = file.Close() }()
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return []packageAnalyzeViolation{}, fmt.Errorf("Error reading %s file: %w", filepath.Base(path), err)
	}

	var result analyzeResultJson
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return []packageAnalyzeViolation{}, fmt.Errorf("Error parsing %s file: %w", filepath.Base(path), err)
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
	governanceFileParam := c.getStringParameter("governance-file", "", context.Parameters)
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

func (c PackageAnalyzeCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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
