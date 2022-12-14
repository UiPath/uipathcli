package commandline

import (
	"fmt"
	"io"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
	"github.com/urfave/cli/v2"
)

type Cli struct {
	StdIn          io.Reader
	StdOut         io.Writer
	StdErr         io.Writer
	ColoredOutput  bool
	Parser         parser.Parser
	ConfigProvider config.ConfigProvider
	Executor       executor.Executor
}

func (c Cli) parseDefinitions(definitions []DefinitionData) ([]parser.Definition, error) {
	result := []parser.Definition{}
	for _, definition := range definitions {
		d, err := c.Parser.Parse(definition.Name, definition.Data)
		if err != nil {
			return nil, fmt.Errorf("Error parsing definition file '%s': %v", definition.Name, err)
		}
		result = append(result, *d)
	}
	return result, nil
}

func (c Cli) run(args []string, configData []byte, definitionData []DefinitionData) error {
	err := c.ConfigProvider.Load(configData)
	if err != nil {
		return err
	}
	definitions, err := c.parseDefinitions(definitionData)
	if err != nil {
		return err
	}

	CommandBuilder := CommandBuilder{
		StdIn:          c.StdIn,
		StdOut:         c.StdOut,
		ConfigProvider: c.ConfigProvider,
		Executor:       c.Executor,
	}
	flags := CommandBuilder.CreateDefaultFlags(false)
	commands := CommandBuilder.Create(definitions)

	app := &cli.App{
		Name:            "uipathcli",
		Usage:           "Command-Line Interface for UiPath Services",
		UsageText:       "uipathcli <service> <operation> --parameter",
		Version:         "1.0.0",
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

func (c Cli) Run(args []string, configData []byte, definitionData []DefinitionData) error {
	err := c.run(args, configData, definitionData)
	if err != nil {
		message := err.Error()
		if c.ColoredOutput {
			message = colorRed + err.Error() + colorReset
		}
		fmt.Fprintln(c.StdErr, message)
	}
	return err
}
