package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/commandline"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/plugins"
)

const DefinitionsDirectory = "definitions"

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

func readConfiguration() (string, []byte, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", nil, fmt.Errorf("Error reading configuration file: %v", err)
	}
	filename := os.Getenv("UIPATHCLI_CONFIGURATION_PATH")
	if filename == "" {
		filename = filepath.Join(homeDir, ".uipathcli", "config")
	}

	data, err := os.ReadFile(filename)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return filename, []byte{}, nil
	}
	if err != nil {
		return "", nil, fmt.Errorf("Error reading configuration file '%s': %v", filename, err)
	}
	return filename, data, nil
}

func readPlugins() (*plugins.Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("Error reading plugins file: %v", err)
	}

	filename := os.Getenv("UIPATHCLI_PLUGINS_PATH")
	if filename == "" {
		filename = filepath.Join(homeDir, ".uipathcli", "plugins")
	}

	data, err := os.ReadFile(filename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("Error reading plugins file '%s': %v", filename, err)
	}
	config_provider := plugins.ConfigProvider{}
	return config_provider.Parse(data)
}

func authenticators(pluginsCfg *plugins.Config) []auth.Authenticator {
	authenticators := []auth.Authenticator{}
	for _, authenticator := range pluginsCfg.Authenticators {
		authenticators = append(authenticators, auth.ExternalAuthenticator{
			Config: *auth.NewExternalAuthenticatorConfig(authenticator.Name, authenticator.Path),
		})
	}
	return append(authenticators,
		auth.PatAuthenticator{},
		auth.OAuthAuthenticator{
			Cache: cache.FileCache{},
		},
		auth.BearerAuthenticator{
			Cache: cache.FileCache{},
		},
	)
}

func colorsSupported() bool {
	_, noColorSet := os.LookupEnv("NO_COLOR")
	term, _ := os.LookupEnv("TERM")
	omitColors := noColorSet || term == "dumb" || runtime.GOOS == "windows"
	return !omitColors
}

func main() {
	cfgFile, cfgData, err := readConfiguration()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(131)
	}
	definitions, err := readDefinitions()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(132)
	}
	pluginsCfg, err := readPlugins()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(133)
	}

	cli := commandline.Cli{
		StdIn:         os.Stdin,
		StdOut:        os.Stdout,
		StdErr:        os.Stderr,
		Parser:        parser.OpenApiParser{},
		ColoredOutput: colorsSupported(),
		ConfigProvider: config.ConfigProvider{
			ConfigFileName: cfgFile,
		},
		Executor: executor.HttpExecutor{
			Authenticators: authenticators(pluginsCfg),
		},
	}

	err = cli.Run(os.Args, cfgData, definitions)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
