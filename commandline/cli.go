// Package commandline is responsible for creating, parsing and validating
// command line arguments.
package commandline

import (
	"fmt"
	"io"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/utils/stream"
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

func (c Cli) run(args []string, input stream.Stream) error {
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

	flags := NewFlagBuilder().
		AddDefaultFlags(false).
		AddVersionFlag().
		Build()

	commands, err := CommandBuilder.Create(args)
	if err != nil {
		return err
	}

	app := &cli.App{
		Name:                      "uipath",
		Usage:                     "Command-Line Interface for UiPath Services",
		UsageText:                 "uipath <service> <operation> --parameter",
		Version:                   "1.0",
		Flags:                     c.convertFlags(flags...),
		Commands:                  c.convertCommands(commands...),
		Writer:                    c.stdOut,
		ErrWriter:                 c.stdErr,
		HideVersion:               true,
		HideHelpCommand:           true,
		DisableSliceFlagSeparator: true,
		Action: func(context *cli.Context) error {
			if context.IsSet(FlagNameVersion) {
				handler := newVersionCommandHandler(c.stdOut)
				return handler.Execute()
			}
			return cli.ShowAppHelp(context)
		},
	}
	return app.Run(args)
}

const colorRed = "\033[31m"
const colorReset = "\033[0m"

func (c Cli) Run(args []string, input stream.Stream) error {
	err := c.run(args, input)
	if err != nil {
		message := err.Error()
		if c.coloredOutput {
			message = colorRed + err.Error() + colorReset
		}
		_, _ = fmt.Fprintln(c.stdErr, message)
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

func (c Cli) convertCommand(command *CommandDefinition) *cli.Command {
	result := cli.Command{
		Name:               command.Name,
		Usage:              command.Summary,
		Description:        command.Description,
		Flags:              c.convertFlags(command.Flags...),
		Subcommands:        c.convertCommands(command.Subcommands...),
		CustomHelpTemplate: command.HelpTemplate,
		Hidden:             command.Hidden,
		HideHelp:           true,
	}
	if command.Action != nil {
		result.Action = func(context *cli.Context) error {
			return command.Action(&CommandExecContext{context})
		}
	}
	return &result
}

func (c Cli) convertCommands(commands ...*CommandDefinition) []*cli.Command {
	result := []*cli.Command{}
	for _, command := range commands {
		result = append(result, c.convertCommand(command))
	}
	return result
}

func (c Cli) convertStringSliceFlag(flag *FlagDefinition) *cli.StringSliceFlag {
	envVars := []string{}
	if flag.EnvVarName != "" {
		envVars = append(envVars, flag.EnvVarName)
	}
	var value *cli.StringSlice
	if flag.DefaultValue != nil {
		value = cli.NewStringSlice(flag.DefaultValue.([]string)...)
	}
	return &cli.StringSliceFlag{
		Name:     flag.Name,
		Usage:    flag.Summary,
		EnvVars:  envVars,
		Required: flag.Required,
		Hidden:   flag.Hidden,
		Value:    value,
	}
}

func (c Cli) convertIntFlag(flag *FlagDefinition) *cli.IntFlag {
	envVars := []string{}
	if flag.EnvVarName != "" {
		envVars = append(envVars, flag.EnvVarName)
	}
	var value int
	if flag.DefaultValue != nil {
		value = flag.DefaultValue.(int)
	}
	return &cli.IntFlag{
		Name:     flag.Name,
		Usage:    flag.Summary,
		EnvVars:  envVars,
		Required: flag.Required,
		Hidden:   flag.Hidden,
		Value:    value,
	}
}

func (c Cli) convertBoolFlag(flag *FlagDefinition) *cli.BoolFlag {
	envVars := []string{}
	if flag.EnvVarName != "" {
		envVars = append(envVars, flag.EnvVarName)
	}
	var value bool
	if flag.DefaultValue != nil {
		value = flag.DefaultValue.(bool)
	}
	return &cli.BoolFlag{
		Name:     flag.Name,
		Usage:    flag.Summary,
		EnvVars:  envVars,
		Required: flag.Required,
		Hidden:   flag.Hidden,
		Value:    value,
	}
}

func (c Cli) convertStringFlag(flag *FlagDefinition) *cli.StringFlag {
	envVars := []string{}
	if flag.EnvVarName != "" {
		envVars = append(envVars, flag.EnvVarName)
	}
	var value string
	if flag.DefaultValue != nil {
		value = flag.DefaultValue.(string)
	}
	return &cli.StringFlag{
		Name:     flag.Name,
		Usage:    flag.Summary,
		EnvVars:  envVars,
		Required: flag.Required,
		Hidden:   flag.Hidden,
		Value:    value,
	}
}

func (c Cli) convertFlag(flag *FlagDefinition) cli.Flag {
	switch flag.Type {
	case FlagTypeStringArray:
		return c.convertStringSliceFlag(flag)
	case FlagTypeInteger:
		return c.convertIntFlag(flag)
	case FlagTypeBoolean:
		return c.convertBoolFlag(flag)
	case FlagTypeString:
		return c.convertStringFlag(flag)
	}
	panic("Unknown flag type: " + flag.Type.String())
}

func (c Cli) convertFlags(flags ...*FlagDefinition) []cli.Flag {
	result := []cli.Flag{}
	for _, flag := range flags {
		result = append(result, c.convertFlag(flag))
	}
	return result
}
