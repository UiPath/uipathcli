package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/UiPath/uipathcli/commandline"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
)

const DefinitionsDirectory = "definitions"
const ConfigurationFileName = "uipathcli.profiles.yaml"

func readDefinition(path string) (*commandline.DefinitionData, error) {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading definition file '%s': %v", path, err)
	}
	return commandline.NewDefinitionData(name, data), nil
}

func readDefinitions() ([]commandline.DefinitionData, error) {
	currentDirectory, err := os.Executable()
	definitionsDirectory := filepath.Join(filepath.Dir(currentDirectory), DefinitionsDirectory)
	if err != nil {
		return nil, fmt.Errorf("Error reading definition files from folder '%s': %v", definitionsDirectory, err)
	}
	files, err := os.ReadDir(definitionsDirectory)
	if err != nil {
		return nil, fmt.Errorf("Error reading definition files from folder '%s': %v", definitionsDirectory, err)
	}

	result := []commandline.DefinitionData{}
	for _, file := range files {
		path := filepath.Join(definitionsDirectory, file.Name())
		data, err := readDefinition(path)
		if err != nil {
			return nil, err
		}
		result = append(result, *data)
	}
	return result, nil
}

func readConfiguration() ([]byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration file: %v", err)
	}
	filename := filepath.Join(homeDir, ConfigurationFileName)
	data, err := os.ReadFile(filename)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return []byte{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("Error reading configuration file '%s': %v", filename, err)
	}
	return data, nil
}

func main() {
	cfg, err := readConfiguration()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(131)
	}
	definitions, err := readDefinitions()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(132)
	}
	cli := commandline.Cli{
		StdOut:         os.Stdout,
		StdErr:         os.Stderr,
		Parser:         parser.OpenApiParser{},
		ConfigProvider: config.ConfigProvider{},
		Executor: executor.HttpExecutor{
			TokenProvider: executor.IdentityClient{
				Cache: executor.FileCache{},
			},
		},
	}

	err = cli.Run(os.Args, cfg, definitions)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
