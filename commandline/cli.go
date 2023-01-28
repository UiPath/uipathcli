package commandline

import (
	"fmt"
	"io"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/plugin"
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
	PluginExecutor executor.Executor
	CommandPlugins []plugin.CommandPlugin
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

func (c Cli) findDefinition(name string, definitions []parser.Definition) *parser.Definition {
	for i := range definitions {
		if definitions[i].Name == name {
			return &definitions[i]
		}
	}
	return nil
}

func (c Cli) convertToParameters(parameters []plugin.CommandParameter) []parser.Parameter {
	result := []parser.Parameter{}
	for _, p := range parameters {
		parameter := *parser.NewParameter(
			p.Name,
			p.Type,
			p.Description,
			parser.ParameterInBody,
			p.Name,
			p.Required,
			nil,
			[]parser.Parameter{})
		result = append(result, parameter)
	}
	return result
}

func (c Cli) applyPluginCommand(plugin plugin.CommandPlugin, command plugin.Command, definition *parser.Definition) {
	parameters := c.convertToParameters(command.Parameters)
	operation := parser.NewOperation(command.Name, command.Description, "", "", "application/json", parameters, plugin, command.Hidden)
	for i, _ := range definition.Operations {
		if definition.Operations[i].Name == command.Name {
			definition.Operations[i] = *operation
			return
		}
	}
	definition.Operations = append(definition.Operations, *operation)
}

func (c Cli) applyPlugins(definitions []parser.Definition) {
	for _, plugin := range c.CommandPlugins {
		command := plugin.Command()
		definition := c.findDefinition(command.Service, definitions)
		if definition != nil {
			c.applyPluginCommand(plugin, command, definition)
		}
	}
}

func (c Cli) run(args []string, configData []byte, definitionData []DefinitionData, input []byte) error {
	err := c.ConfigProvider.Load(configData)
	if err != nil {
		return err
	}
	definitions, err := c.parseDefinitions(definitionData)
	if err != nil {
		return err
	}
	c.applyPlugins(definitions)

	CommandBuilder := CommandBuilder{
		Input:          input,
		StdIn:          c.StdIn,
		StdOut:         c.StdOut,
		ConfigProvider: c.ConfigProvider,
		Executor:       c.Executor,
		PluginExecutor: c.PluginExecutor,
	}
	flags := CommandBuilder.CreateDefaultFlags(false)
	commands := CommandBuilder.Create(definitions)

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

func (c Cli) Run(args []string, configData []byte, definitionData []DefinitionData, input []byte) error {
	err := c.run(args, configData, definitionData, input)
	if err != nil {
		message := err.Error()
		if c.ColoredOutput {
			message = colorRed + err.Error() + colorReset
		}
		fmt.Fprintln(c.StdErr, message)
	}
	return err
}
