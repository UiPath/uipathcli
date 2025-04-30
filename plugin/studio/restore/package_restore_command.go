package restore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/utils/process"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The PackageRestoreCommand restores the packages of the project
type PackageRestoreCommand struct {
	Exec process.ExecProcess
}

func (c PackageRestoreCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("package", "Package", "UiPath Studio package-related actions").
		WithOperation("restore", "Package Project", "Restores the packages of the project").
		WithParameter(plugin.NewParameter("source", plugin.ParameterTypeString, "Path to a project.json file or a folder containing project.json file").
			WithRequired(true).
			WithDefaultValue(".")).
		WithParameter(plugin.NewParameter("destination", plugin.ParameterTypeString, "The output folder").
			WithRequired(true).
			WithDefaultValue("./packages"))
}

func (c PackageRestoreCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	source, err := c.getSource(ctx)
	if err != nil {
		return err
	}
	destination := c.getDestination(ctx)

	params := newPackageRestoreParams(
		ctx.Organization,
		ctx.Tenant,
		ctx.BaseUri,
		ctx.Auth.Token,
		ctx.IdentityUri,
		source,
		destination)

	result, err := c.execute(*params, ctx.Debug, logger)
	if err != nil {
		return err
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("restore command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(json)))
}

func (c PackageRestoreCommand) execute(params packageRestoreParams, debug bool, logger log.Logger) (*packageRestoreResult, error) {
	projectReader := studio.NewStudioProjectReader(params.Source)
	project, err := projectReader.ReadMetadata()
	if err != nil {
		return nil, err
	}
	supported, err := project.TargetFramework.IsSupported()
	if !supported {
		return nil, err
	}

	uipcli := studio.NewUipcli(c.Exec, logger)
	err = uipcli.Initialize(project.TargetFramework)
	if err != nil {
		return nil, err
	}

	if !debug {
		bar := c.newProgressBar(logger)
		defer close(bar)
	}
	args := c.prepareRestoreArguments(params)
	exitCode, stdErr, err := uipcli.ExecuteAndWait(args...)
	if err != nil {
		return nil, err
	}

	var result *packageRestoreResult
	if exitCode == 0 {
		if err != nil {
			return nil, err
		}
		result = newSucceededPackageRestoreResult(
			params.Destination,
			project.Name,
			project.Description,
			project.ProjectId)
	} else {
		result = newFailedPackageRestoreResult(
			stdErr,
			&project.Name,
			&project.Description,
			&project.ProjectId)
	}
	return result, nil
}

func (c PackageRestoreCommand) prepareRestoreArguments(params packageRestoreParams) []string {
	source, _ := strings.CutSuffix(params.Source, studio.DefaultProjectJson)
	args := []string{"package", "restore", source, "--restoreFolder", params.Destination}
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

func (c PackageRestoreCommand) newProgressBar(logger log.Logger) chan struct{} {
	progressBar := visualization.NewProgressBar(logger)
	ticker := time.NewTicker(10 * time.Millisecond)
	cancel := make(chan struct{})
	var percent float64 = 0
	go func() {
		for {
			select {
			case <-ticker.C:
				progressBar.UpdatePercentage("restoring...  ", percent)
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

func (c PackageRestoreCommand) getSource(ctx plugin.ExecutionContext) (string, error) {
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

func (c PackageRestoreCommand) getDestination(ctx plugin.ExecutionContext) string {
	destination := c.getStringParameter("destination", "./packages/", ctx.Parameters)
	destination, _ = filepath.Abs(destination)
	return destination
}

func (c PackageRestoreCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func NewPackageRestoreCommand() *PackageRestoreCommand {
	return &PackageRestoreCommand{process.NewExecProcess()}
}
