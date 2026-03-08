// Package pack implements the command plugin for packing a UiPath solution
// directory into a .uis file (ZIP archive).
package pack

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

// The SolutionPackCommand packs a solution directory into a .uis file.
type SolutionPackCommand struct{}

func (c SolutionPackCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("solution", "UiPath Solution management", "Pack, unpack, push and pull UiPath Maestro solutions.").
		WithOperation("pack", "Pack Solution", "Packs a solution directory into a .uis file").
		WithParameter(plugin.NewParameter("source", plugin.ParameterTypeString, "Path to solution directory").
			WithRequired(true).
			WithDefaultValue(".")).
		WithParameter(plugin.NewParameter("destination", plugin.ParameterTypeString, "Output .uis file path").
			WithDefaultValue(""))
}

func (c SolutionPackCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	source := c.getStringParameter("source", ".", ctx.Parameters)
	source, _ = filepath.Abs(source)
	destination := c.getStringParameter("destination", "", ctx.Parameters)

	fileInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("Solution directory not found: %s", source)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("Source is not a directory: %s", source)
	}

	solutionStoragePath := filepath.Join(source, "SolutionStorage.json")
	if _, err := os.Stat(solutionStoragePath); err != nil {
		return fmt.Errorf("SolutionStorage.json not found in %s. This does not appear to be a valid UiPath solution directory", source)
	}

	if destination == "" {
		destination = filepath.Base(source) + ".uis"
	}
	destination, _ = filepath.Abs(destination)

	solutionId, solutionName := c.readSolutionInfo(solutionStoragePath)

	params := newSolutionPackParams(source, destination, solutionId, solutionName)
	result, err := c.pack(*params)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Pack command failed: %w", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(http.StatusOK, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(jsonData)))
}

func (c SolutionPackCommand) pack(params solutionPackParams) (*solutionPackResult, error) {
	outFile, err := os.Create(params.Destination)
	if err != nil {
		return nil, fmt.Errorf("Cannot create output file: %w", err)
	}
	defer func() { _ = outFile.Close() }()

	zipWriter := zip.NewWriter(outFile)
	defer func() { _ = zipWriter.Close() }()

	err = filepath.Walk(params.Source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(params.Source, path)
		if err != nil {
			return err
		}

		// Skip .git directory
		if strings.HasPrefix(relPath, ".git"+string(filepath.Separator)) || relPath == ".git" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip __pycache__ directories
		if info.IsDir() && info.Name() == "__pycache__" {
			return filepath.SkipDir
		}

		// Skip .pyc files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".pyc") {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Use forward slashes in ZIP entries
		zipPath := strings.ReplaceAll(relPath, string(filepath.Separator), "/")
		w, err := zipWriter.Create(zipPath)
		if err != nil {
			return fmt.Errorf("Error creating zip entry '%s': %w", zipPath, err)
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Error opening file '%s': %w", path, err)
		}
		defer func() { _ = f.Close() }()

		_, err = io.Copy(w, f)
		if err != nil {
			return fmt.Errorf("Error writing file '%s': %w", zipPath, err)
		}
		return nil
	})
	if err != nil {
		// Clean up partial output on failure
		_ = zipWriter.Close()
		_ = outFile.Close()
		_ = os.Remove(params.Destination)
		return nil, err
	}

	fileInfo, err := os.Stat(params.Destination)
	size := int64(0)
	if err == nil {
		size = fileInfo.Size()
	}

	return newSucceededSolutionPackResult(params.Destination, params.SolutionId, params.SolutionName, size), nil
}

func (c SolutionPackCommand) readSolutionInfo(path string) (string, string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ""
	}
	var storage struct {
		SolutionId string `json:"SolutionId"`
	}
	err = json.Unmarshal(data, &storage)
	if errors.Is(err, nil) {
		return storage.SolutionId, ""
	}
	return "", ""
}

func (c SolutionPackCommand) getStringParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
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

func NewSolutionPackCommand() *SolutionPackCommand {
	return &SolutionPackCommand{}
}
