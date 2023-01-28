package main

import (
	"errors"
	"fmt"
	"io"
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
	"github.com/UiPath/uipathcli/plugin"
	plugin_digitizer "github.com/UiPath/uipathcli/plugin/digitizer"
)

const DefinitionsDirectory = "definitions"

func readDefinition(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading definition file '%s': %v", path, err)
	}
	return data, nil
}

func definitionsPath() (string, error) {
	path := os.Getenv("UIPATHCLI_DEFINITIONS_PATH")
	if path != "" {
		return path, nil
	}
	currentDirectory, err := os.Executable()
	definitionsDirectory := filepath.Join(filepath.Dir(currentDirectory), DefinitionsDirectory)
	if err != nil {
		return "", fmt.Errorf("Error reading definition files from folder '%s': %v", definitionsDirectory, err)
	}
	return definitionsDirectory, nil
}

func readDefinitions(definitionName string) ([]commandline.DefinitionData, error) {
	definitionsDirectory, err := definitionsPath()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(definitionsDirectory)
	if err != nil {
		return nil, fmt.Errorf("Error reading definition files from folder '%s': %v", definitionsDirectory, err)
	}

	result := []commandline.DefinitionData{}
	for _, file := range files {
		path := filepath.Join(definitionsDirectory, file.Name())
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		data := commandline.NewDefinitionData(name, []byte{})
		if definitionName == name {
			content, err := readDefinition(path)
			if err != nil {
				return nil, err
			}
			data = commandline.NewDefinitionData(name, content)
		}
		result = append(result, *data)
	}
	return result, nil
}

func configurationFilePath() (string, error) {
	filename := os.Getenv("UIPATHCLI_CONFIGURATION_PATH")
	if filename != "" {
		return filename, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error reading configuration file: %v", err)
	}
	filename = filepath.Join(homeDir, ".uipathcli", "config")
	return filename, nil
}

func readConfiguration() (string, []byte, error) {
	filename, err := configurationFilePath()
	if err != nil {
		return "", nil, err
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

func pluginsFilePath() (string, error) {
	filename := os.Getenv("UIPATHCLI_PLUGINS_PATH")
	if filename != "" {
		return filename, nil
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("Error reading plugins file: %v", err)
	}
	filename = filepath.Join(homeDir, ".uipathcli", "plugins")
	return filename, nil
}

func readPlugins() (*config.PluginConfig, error) {
	filename, err := pluginsFilePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("Error reading plugins file '%s': %v", filename, err)
	}
	config_provider := config.PluginConfigProvider{}
	return config_provider.Parse(data)
}

func authenticators(pluginsCfg *config.PluginConfig) []auth.Authenticator {
	authenticators := []auth.Authenticator{}
	for _, authenticator := range pluginsCfg.Authenticators {
		authenticators = append(authenticators, auth.ExternalAuthenticator{
			Config: *auth.NewExternalAuthenticatorConfig(authenticator.Name, authenticator.Path),
		})
	}
	return append(authenticators,
		auth.PatAuthenticator{},
		auth.OAuthAuthenticator{
			Cache:           cache.FileCache{},
			BrowserLauncher: auth.ExecBrowserLauncher{},
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

func readStdIn() []byte {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		input, err := io.ReadAll(os.Stdin)
		if err == nil {
			return input
		}
	}
	return []byte{}
}

func requiredDefinition(args []string) string {
	if len(args) <= 1 {
		return ""
	}
	if strings.HasPrefix(args[1], "--") {
		return ""
	}
	if len(args) == 5 && args[1] == "autocomplete" && args[2] == "complete" && args[3] == "--command" {
		autocompleteArgs := strings.Split(args[4], " ")
		return requiredDefinition(autocompleteArgs)
	}
	return args[1]
}

func main() {
	cfgFile, cfgData, err := readConfiguration()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(131)
	}
	definitionName := requiredDefinition(os.Args)
	definitions, err := readDefinitions(definitionName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(132)
	}
	pluginsCfg, err := readPlugins()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(133)
	}

	authenticators := authenticators(pluginsCfg)
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
			Authenticators: authenticators,
		},
		PluginExecutor: executor.PluginExecutor{
			Authenticators: authenticators,
		},
		CommandPlugins: []plugin.CommandPlugin{
			plugin_digitizer.DigitizeCommand{},
			plugin_digitizer.StatusCommand{},
		},
	}

	input := readStdIn()
	err = cli.Run(os.Args, cfgData, definitions, input)
	if err != nil {
		os.Exit(1)
	}
}
