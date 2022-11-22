package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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
	authenticators = append(authenticators, auth.BearerAuthenticator{
		Cache: cache.FileCache{},
	})

	return authenticators
}

func readUserInput(message string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message + " ")
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Trim(value, " \r\n\t"), nil
}

func configureCli(configFile string, configData []byte) error {
	configProvider := config.ConfigProvider{}
	err := configProvider.Load(configData)
	if err != nil {
		return err
	}

	clientId, err := readUserInput("Enter client id:")
	if err != nil {
		return nil
	}
	clientSecret, err := readUserInput("Enter client secret:")
	if err != nil {
		return nil
	}
	organization, err := readUserInput("Enter organization:")
	if err != nil {
		return nil
	}
	tenant, err := readUserInput("Enter tenant:")
	if err != nil {
		return nil
	}

	data, err := configProvider.Update(clientId, clientSecret, organization, tenant)
	if err != nil {
		return err
	}
	os.WriteFile(configFile, data, 0600)
	if err != nil {
		return err
	}
	fmt.Println("Successfully configured uipathcli")
	return nil
}

func main() {
	cfgFile, cfgData, err := readConfiguration()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(131)
	}
	if len(os.Args) > 1 && os.Args[1] == "config" {
		err := configureCli(cfgFile, cfgData)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(140)
		}
		os.Exit(0)
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
		StdOut:         os.Stdout,
		StdErr:         os.Stderr,
		Parser:         parser.OpenApiParser{},
		ConfigProvider: config.ConfigProvider{},
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
