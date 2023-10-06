package commandline

import (
	"encoding/json"
	"sort"

	"github.com/UiPath/uipathcli/parser"
	"github.com/urfave/cli/v2"
)

// showCommandHandler prints all available uipathcli commands
type showCommandHandler struct {
}

type parameterJson struct {
	Name          string        `json:"name"`
	Type          string        `json:"type"`
	Description   string        `json:"description"`
	Required      bool          `json:"required"`
	AllowedValues []interface{} `json:"allowedValues"`
	DefaultValue  interface{}   `json:"defaultValue"`
	Example       string        `json:"example"`
}

type commandJson struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  []parameterJson `json:"parameters"`
	Subcommands []commandJson   `json:"subcommands"`
}

func (h showCommandHandler) Execute(definitions []parser.Definition, globalFlags []cli.Flag) (string, error) {
	result := commandJson{
		Name:        "uipath",
		Description: "Command line interface to simplify, script and automate API calls for UiPath services",
		Parameters:  h.convertFlagsToCommandParameters(globalFlags),
		Subcommands: h.convertDefinitionsToCommands(definitions),
	}
	bytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h showCommandHandler) convertDefinitionsToCommands(definitions []parser.Definition) []commandJson {
	commands := []commandJson{}
	for _, d := range definitions {
		command := h.convertDefinitionToCommands(d)
		commands = append(commands, command)
	}
	return commands
}

func (h showCommandHandler) convertDefinitionToCommands(definition parser.Definition) commandJson {
	categories := map[string]commandJson{}

	for _, op := range definition.Operations {
		if op.Category == nil {
			command := h.convertOperationToCommand(op)
			categories[command.Name] = command
		} else {
			h.createOrUpdateCategory(op, categories)
		}
	}

	commands := []commandJson{}
	for _, command := range categories {
		commands = append(commands, command)
	}

	h.sort(commands)
	for _, command := range commands {
		h.sort(command.Subcommands)
	}
	return commandJson{
		Name:        definition.Name,
		Description: definition.Description,
		Subcommands: commands,
	}
}

func (h showCommandHandler) createOrUpdateCategory(operation parser.Operation, categories map[string]commandJson) {
	command, found := categories[operation.Category.Name]
	if !found {
		command = h.createCategoryCommand(operation)
	}
	command.Subcommands = append(command.Subcommands, h.convertOperationToCommand(operation))
	categories[operation.Category.Name] = command
}

func (h showCommandHandler) createCategoryCommand(operation parser.Operation) commandJson {
	return commandJson{
		Name:        operation.Category.Name,
		Description: operation.Category.Description,
	}
}

func (h showCommandHandler) convertOperationToCommand(operation parser.Operation) commandJson {
	return commandJson{
		Name:        operation.Name,
		Description: operation.Description,
		Parameters:  h.convertParametersToCommandParameters(operation.Parameters),
	}
}

func (h showCommandHandler) convertFlagsToCommandParameters(flags []cli.Flag) []parameterJson {
	result := []parameterJson{}
	for _, f := range flags {
		result = append(result, h.convertFlagToCommandParameter(f))
	}
	return result
}

func (h showCommandHandler) convertParametersToCommandParameters(parameters []parser.Parameter) []parameterJson {
	result := []parameterJson{}
	for _, p := range parameters {
		result = append(result, h.convertParameterToCommandParameter(p))
	}
	return result
}

func (h showCommandHandler) convertFlagToCommandParameter(flag cli.Flag) parameterJson {
	intFlag, ok := flag.(*cli.IntFlag)
	if ok {
		return parameterJson{
			Name:          intFlag.Name,
			Description:   intFlag.Usage,
			Type:          "integer",
			Required:      false,
			AllowedValues: []interface{}{},
			DefaultValue:  intFlag.Value,
		}
	}
	boolFlag, ok := flag.(*cli.BoolFlag)
	if ok {
		return parameterJson{
			Name:          boolFlag.Name,
			Description:   boolFlag.Usage,
			Type:          "boolean",
			Required:      false,
			AllowedValues: []interface{}{},
			DefaultValue:  boolFlag.Value,
		}
	}
	stringFlag := flag.(*cli.StringFlag)
	return parameterJson{
		Name:          stringFlag.Name,
		Description:   stringFlag.Usage,
		Type:          "string",
		Required:      false,
		AllowedValues: []interface{}{},
		DefaultValue:  stringFlag.Value,
	}
}

func (h showCommandHandler) convertParameterToCommandParameter(parameter parser.Parameter) parameterJson {
	formatter := newParameterFormatter(parameter)
	return parameterJson{
		Name:          parameter.Name,
		Description:   parameter.Description,
		Type:          parameter.Type,
		Required:      parameter.Required,
		AllowedValues: parameter.AllowedValues,
		DefaultValue:  parameter.DefaultValue,
		Example:       formatter.UsageExample(),
	}
}

func (h showCommandHandler) sort(commands []commandJson) {
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
}

func newShowCommandHandler() *showCommandHandler {
	return &showCommandHandler{}
}
