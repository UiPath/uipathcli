// Package main contains the entry point of the uipath command line interface.
//
// It only initializes the different packages and delegates the actual work.
package main

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/commandline"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/plugin"
	plugin_digitizer "github.com/UiPath/uipathcli/plugin/digitizer"
	plugin_orchestrator "github.com/UiPath/uipathcli/plugin/orchestrator"
)

func authenticators(pluginsCfg config.PluginConfig) []auth.Authenticator {
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

func main() {
	configProvider := config.ConfigProvider{
		ConfigStore: config.ConfigStore{
			ConfigFile: os.Getenv("UIPATH_CONFIGURATION_PATH"),
		},
	}
	err := configProvider.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(131)
	}
	pluginConfigProvider := config.PluginConfigProvider{
		PluginConfigStore: config.PluginConfigStore{
			PluginFile: os.Getenv("UIPATH_PLUGINS_PATH"),
		},
	}
	err = pluginConfigProvider.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(132)
	}

	pluginConfig := pluginConfigProvider.Config()
	authenticators := authenticators(pluginConfig)
	cli := commandline.Cli{
		StdIn:         os.Stdin,
		StdOut:        os.Stdout,
		StdErr:        os.Stderr,
		ColoredOutput: colorsSupported(),
		DefinitionProvider: commandline.DefinitionProvider{
			DefinitionStore: commandline.DefinitionStore{
				DefinitionDirectory: os.Getenv("UIPATH_DEFINITIONS_PATH"),
			},
			Parser: parser.OpenApiParser{},
			CommandPlugins: []plugin.CommandPlugin{
				plugin_digitizer.DigitizeCommand{},
				plugin_orchestrator.UploadCommand{},
				plugin_orchestrator.DownloadCommand{},
			},
		},
		ConfigProvider: configProvider,
		Executor: executor.HttpExecutor{
			Authenticators: authenticators,
		},
		PluginExecutor: executor.PluginExecutor{
			Authenticators: authenticators,
		},
	}

	input := readStdIn()
	err = cli.Run(os.Args, input)
	if err != nil {
		os.Exit(1)
	}
}
