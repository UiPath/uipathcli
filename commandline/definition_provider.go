package commandline

import (
	"fmt"

	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/plugin"
)

type DefinitionProvider struct {
	DefinitionStore DefinitionStore
	Parser          parser.Parser
	CommandPlugins  []plugin.CommandPlugin
}

func (p DefinitionProvider) Index() ([]parser.Definition, error) {
	emptyDefinitions, err := p.loadEmptyDefinitions()
	if err != nil {
		return nil, err
	}
	return p.parse(emptyDefinitions)
}

func (p DefinitionProvider) Load(name string) ([]parser.Definition, error) {
	data, err := p.loadDefinitionWithData(name)
	if err != nil {
		return nil, err
	}
	return p.parse(data)
}

func (p DefinitionProvider) loadEmptyDefinitions() ([]DefinitionData, error) {
	names, err := p.DefinitionStore.Names()
	if err != nil {
		return nil, err
	}
	result := []DefinitionData{}
	for _, name := range names {
		result = append(result, *NewDefinitionData(name, []byte{}))
	}
	return result, nil
}

func (p DefinitionProvider) loadDefinitionWithData(name string) ([]DefinitionData, error) {
	definition, err := p.DefinitionStore.Read(name)
	if err != nil {
		return nil, err
	}
	if definition != nil {
		return []DefinitionData{*definition}, nil
	}
	return []DefinitionData{}, nil
}

func (p DefinitionProvider) parse(data []DefinitionData) ([]parser.Definition, error) {
	definitions, err := p.parseDefinitions(data)
	if err != nil {
		return nil, err
	}
	p.applyPlugins(definitions)
	return definitions, nil
}

func (p DefinitionProvider) parseDefinitions(definitions []DefinitionData) ([]parser.Definition, error) {
	result := []parser.Definition{}
	for _, definition := range definitions {
		d, err := p.Parser.Parse(definition.Name, definition.Data)
		if err != nil {
			return nil, fmt.Errorf("Error parsing definition file '%s': %v", definition.Name, err)
		}
		result = append(result, *d)
	}
	return result, nil
}

func (p DefinitionProvider) findDefinition(name string, definitions []parser.Definition) *parser.Definition {
	for i := range definitions {
		if definitions[i].Name == name {
			return &definitions[i]
		}
	}
	return nil
}

func (p DefinitionProvider) convertToParameters(parameters []plugin.CommandParameter) []parser.Parameter {
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

func (p DefinitionProvider) applyPluginCommand(plugin plugin.CommandPlugin, command plugin.Command, definition *parser.Definition) {
	parameters := p.convertToParameters(command.Parameters)
	operation := parser.NewOperation(command.Name, command.Description, "", "", "application/json", parameters, plugin, command.Hidden)
	for i := range definition.Operations {
		if definition.Operations[i].Name == command.Name {
			definition.Operations[i] = *operation
			return
		}
	}
	definition.Operations = append(definition.Operations, *operation)
}

func (p DefinitionProvider) applyPlugins(definitions []parser.Definition) {
	for _, plugin := range p.CommandPlugins {
		command := plugin.Command()
		definition := p.findDefinition(command.Service, definitions)
		if definition != nil {
			p.applyPluginCommand(plugin, command, definition)
		}
	}
}
