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
		err := c.extractFile(file, params.Destination)
		if err != nil {
			return nil, err
		}
	}

	solutionId, projectCount := c.readSolutionInfo(filepath.Join(params.Destination, "SolutionStorage.json"))

	return newSucceededSolutionUnpackResult(params.Destination, solutionId, projectCount), nil
}

const maxArchiveFileSize = 1 * 1024 * 1024 * 1024

func (c SolutionUnpackCommand) extractFile(file *zip.File, destination string) error {
	destPath, err := c.sanitizeArchivePath(destination, file.Name)
	if err != nil {
		return err
	}

	if file.FileInfo().IsDir() {
		return os.MkdirAll(destPath, 0750)
	}

	err = os.MkdirAll(filepath.Dir(destPath), 0750)
	if err != nil {
		return fmt.Errorf("Cannot create directory for '%s': %w", destPath, err)
	}

	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("Cannot read archive entry '%s': %w", file.Name, err)
	}
	defer func() { _ = rc.Close() }()

	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("Cannot create file '%s': %w", destPath, err)
	}
	defer func() { _ = outFile.Close() }()

	_, err = io.CopyN(outFile, rc, maxArchiveFileSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("Error extracting '%s': %w", file.Name, err)
	}
	return nil
}

func (c SolutionUnpackCommand) sanitizeArchivePath(directory string, name string) (string, error) {
	result := filepath.Join(directory, name)
	if strings.HasPrefix(result, filepath.Clean(directory)) {
		return result, nil
	}
	return "", fmt.Errorf("File path '%s' is not allowed", name)
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
