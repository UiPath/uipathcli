package commandline

import (
	"fmt"
	"io"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/urfave/cli/v2"
)

type Cli struct {
	StdIn              io.Reader
	StdOut             io.Writer
	StdErr             io.Writer
	ColoredOutput      bool
	DefinitionProvider DefinitionProvider
	ConfigProvider     config.ConfigProvider
	Executor           executor.Executor
	PluginExecutor     executor.Executor
}

func (c Cli) run(args []string, input []byte) error {
	err := c.ConfigProvider.Load()
	if err != nil {
		return err
	}

	CommandBuilder := CommandBuilder{
		Input:              input,
		StdIn:              c.StdIn,
		StdOut:             c.StdOut,
		StdErr:             c.StdErr,
		ConfigProvider:     c.ConfigProvider,
		Executor:           c.Executor,
		PluginExecutor:     c.PluginExecutor,
		DefinitionProvider: c.DefinitionProvider,
	}
	flags := CommandBuilder.CreateDefaultFlags(false)
	commands, err := CommandBuilder.Create(args)
	if err != nil {
		return err
	}

	app := &cli.App{
		Name:            "uipathcli",
		Usage:           "Command-Line Interface for UiPath Services",
		UsageText:       "uipathcli <service> <operation> --parameter",
		Version:         "1.0",
		Flags:           flags,
		Commands:        commands,
		Writer:          c.StdOut,
		ErrWriter:       c.StdErr,
		HideVersion:     true,
		HideHelpCommand: true,
	}
	return app.Run(args)
}

const colorRed = "\033[31m"
const colorReset = "\033[0m"

func (c Cli) Run(args []string, input []byte) error {
	err := c.run(args, input)
	if err != nil {
		message := err.Error()
		if c.ColoredOutput {
			message = colorRed + err.Error() + colorReset
		}
		fmt.Fprintln(c.StdErr, message)
	}
	return err
}
