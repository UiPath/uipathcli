// Package unpack implements the command plugin for extracting a .uis file
// (ZIP archive) into a solution directory.
package unpack

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
)

// The SolutionUnpackCommand extracts a .uis file into a directory.
type SolutionUnpackCommand struct{}

func (c SolutionUnpackCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("solution", "UiPath Solution management", "Pack, unpack, push and pull UiPath Maestro solutions.").
		WithOperation("unpack", "Unpack Solution", "Extracts a .uis file into a solution directory").
		WithParameter(plugin.NewParameter("source", plugin.ParameterTypeString, "Path to .uis file").
			WithRequired(true)).
		WithParameter(plugin.NewParameter("destination", plugin.ParameterTypeString, "Output directory path").
			WithDefaultValue(""))
}

func (c SolutionUnpackCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	source := c.getStringParameter("source", "", ctx.Parameters)
	if source == "" {
		return errors.New("Source .uis file is required")
	}
	source, _ = filepath.Abs(source)
	destination := c.getStringParameter("destination", "", ctx.Parameters)

	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("File not found: %s", source)
	}

	if destination == "" {
		basename := filepath.Base(source)
		destination = strings.TrimSuffix(basename, filepath.Ext(basename))
	}
	destination, _ = filepath.Abs(destination)

	params := newSolutionUnpackParams(source, destination)
	result, err := c.unpack(*params)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Unpack command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(jsonData)))
}

func (c SolutionUnpackCommand) unpack(params solutionUnpackParams) (*solutionUnpackResult, error) {
	reader, err := zip.OpenReader(params.Source)
	if err != nil {
		return nil, fmt.Errorf("Cannot open .uis file: %w", err)
	}
	defer func() { _ = reader.Close() }()

	for _, file := range reader.File {
		destPath := filepath.Join(params.Destination, file.Name) //nolint:gosec // paths within trusted .uis archive

		// Validate path doesn't escape destination (zip slip protection)
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(params.Destination)+string(filepath.Separator)) &&
			filepath.Clean(destPath) != filepath.Clean(params.Destination) {
			return nil, fmt.Errorf("Invalid file path in archive: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			err := os.MkdirAll(destPath, 0750)
			if err != nil {
				return nil, fmt.Errorf("Cannot create directory '%s': %w", destPath, err)
			}
			continue
		}

		err := os.MkdirAll(filepath.Dir(destPath), 0750)
		if err != nil {
			return nil, fmt.Errorf("Cannot create directory for '%s': %w", destPath, err)
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			return nil, fmt.Errorf("Cannot create file '%s': %w", destPath, err)
		}

		rc, err := file.Open()
		if err != nil {
			_ = outFile.Close()
			return nil, fmt.Errorf("Cannot read archive entry '%s': %w", file.Name, err)
		}

		const maxFileSize = 1 << 30 // 1 GB
		_, err = io.Copy(outFile, io.LimitReader(rc, maxFileSize))
		_ = rc.Close()
		_ = outFile.Close()
		if err != nil {
			return nil, fmt.Errorf("Error extracting '%s': %w", file.Name, err)
		}
	}

	solutionId, projectCount := c.readSolutionInfo(filepath.Join(params.Destination, "SolutionStorage.json"))

	return newSucceededSolutionUnpackResult(params.Destination, solutionId, projectCount), nil
}

func (c SolutionUnpackCommand) readSolutionInfo(path string) (string, int) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", 0
	}
	var storage struct {
		SolutionId string `json:"SolutionId"`
		Projects   []struct {
			ProjectId string `json:"ProjectId"`
		} `json:"Projects"`
	}
	err = json.Unmarshal(data, &storage)
	if err != nil {
		return "", 0
	}
	return storage.SolutionId, len(storage.Projects)
}

func (c SolutionUnpackCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func NewSolutionUnpackCommand() *SolutionUnpackCommand {
	return &SolutionUnpackCommand{}
}
