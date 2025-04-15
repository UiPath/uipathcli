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
	plugin_orchestrator_download "github.com/UiPath/uipathcli/plugin/orchestrator/download"
	plugin_orchestrator_upload "github.com/UiPath/uipathcli/plugin/orchestrator/upload"
	plugin_studio_analyze "github.com/UiPath/uipathcli/plugin/studio/analyze"
	plugin_studio_pack "github.com/UiPath/uipathcli/plugin/studio/pack"
	plugin_studio_publish "github.com/UiPath/uipathcli/plugin/studio/publish"
	plugin_studio_restore "github.com/UiPath/uipathcli/plugin/studio/restore"
	plugin_studio_testrun "github.com/UiPath/uipathcli/plugin/studio/testrun"
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
				plugin_orchestrator_download.NewDownloadCommand(),
				plugin_orchestrator_upload.NewUploadCommand(),
				plugin_studio_pack.NewPackagePackCommand(),
				plugin_studio_analyze.NewPackageAnalyzeCommand(),
				plugin_studio_restore.NewPackageRestoreCommand(),
				plugin_studio_publish.NewPackagePublishCommand(),
				plugin_studio_testrun.NewTestRunCommand(),
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
