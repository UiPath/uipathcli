// Package commandline is responsible for creating, parsing and validating
// command line arguments.
package commandline

import (
	"fmt"
	"io"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/utils"
	"github.com/urfave/cli/v2"
)

// Cli is a wrapper for building the CLI commands.
type Cli struct {
	stdIn              io.Reader
	stdOut             io.Writer
	stdErr             io.Writer
	coloredOutput      bool
	definitionProvider DefinitionProvider
	configProvider     config.ConfigProvider
	executor           executor.Executor
	pluginExecutor     executor.Executor
}

func (c Cli) run(args []string, input utils.Stream) error {
	err := c.configProvider.Load()
	if err != nil {
		return err
	}

	CommandBuilder := CommandBuilder{
		Input:              input,
		StdIn:              c.stdIn,
		StdOut:             c.stdOut,
		StdErr:             c.stdErr,
		ConfigProvider:     c.configProvider,
		Executor:           c.executor,
		PluginExecutor:     c.pluginExecutor,
		DefinitionProvider: c.definitionProvider,
	}
	flags := CommandBuilder.CreateDefaultFlags(false)
	commands, err := CommandBuilder.Create(args)
	if err != nil {
		return err
	}

	app := &cli.App{
		Name:                      "uipath",
		Usage:                     "Command-Line Interface for UiPath Services",
		UsageText:                 "uipath <service> <operation> --parameter",
		Version:                   "1.0",
		Flags:                     flags,
		Commands:                  commands,
		Writer:                    c.stdOut,
		ErrWriter:                 c.stdErr,
		HideVersion:               true,
		HideHelpCommand:           true,
		DisableSliceFlagSeparator: true,
	}
	return app.Run(args)
}

const colorRed = "\033[31m"
const colorReset = "\033[0m"

func (c Cli) Run(args []string, input utils.Stream) error {
	err := c.run(args, input)
	if err != nil {
		message := err.Error()
		if c.coloredOutput {
			message = colorRed + err.Error() + colorReset
		}
		fmt.Fprintln(c.stdErr, message)
	}
	return err
}

func NewCli(
	stdIn io.Reader,
	stdOut io.Writer,
	stdErr io.Writer,
	colors bool,
	definitionProvider DefinitionProvider,
	configProvider config.ConfigProvider,
	executor executor.Executor,
	pluginExecutor executor.Executor,
) *Cli {
	return &Cli{stdIn, stdOut, stdErr, colors, definitionProvider, configProvider, executor, pluginExecutor}
}
