// Package main contains the entry point of the uipath command line interface.
//
// It only initializes the different packages and delegates the actual work.
package main

import (
	"embed"
	"fmt"
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
	plugin_studio "github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/utils/stream"
)

//go:embed definitions/*.yaml
var embedded embed.FS

func authenticators() []auth.Authenticator {
	return []auth.Authenticator{
		auth.NewPatAuthenticator(),
		auth.NewOAuthAuthenticator(cache.NewFileCache(), *auth.NewBrowserLauncher()),
		auth.NewBearerAuthenticator(cache.NewFileCache()),
	}
}

func colorsSupported() bool {
	_, noColorSet := os.LookupEnv("NO_COLOR")
	term, _ := os.LookupEnv("TERM")
	omitColors := noColorSet || term == "dumb" || runtime.GOOS == "windows"
	return !omitColors
}

func stdIn() stream.Stream {
	return stream.NewReaderStream(parser.RawBodyParameterName, os.Stdin)
}

func main() {
	configProvider := config.NewConfigProvider(
		config.NewConfigFileStore(os.Getenv("UIPATH_CONFIGURATION_PATH")),
	)
	err := configProvider.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(131)
	}

	authenticators := authenticators()
	cli := commandline.NewCli(
		os.Stdin,
		os.Stdout,
		os.Stderr,
		colorsSupported(),
		*commandline.NewDefinitionProvider(
			commandline.NewDefinitionFileStore(os.Getenv("UIPATH_DEFINITIONS_PATH"), embedded),
			parser.NewOpenApiParser(),
			[]plugin.CommandPlugin{
				plugin_digitizer.NewDigitizeCommand(),
				plugin_orchestrator.NewUploadCommand(),
				plugin_orchestrator.NewDownloadCommand(),
				plugin_studio.NewPackagePackCommand(),
				plugin_studio.NewPackageAnalyzeCommand(),
			},
		),
		*configProvider,
		executor.NewHttpExecutor(authenticators),
		executor.NewPluginExecutor(authenticators),
	)

	input := stdIn()
	err = cli.Run(os.Args, input)
	if err != nil {
		os.Exit(1)
	}
}
