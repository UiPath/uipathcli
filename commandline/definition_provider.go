package commandline

import (
	"fmt"
	"net/url"
	"strings"

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
	result := []parser.Definition{}
	for _, data := range emptyDefinitions {
		definition, err := p.parse(data)
		if err != nil {
			return nil, err
		}
		result = append(result, *definition)
	}
	return result, nil
}

func (p DefinitionProvider) Load(name string) (*parser.Definition, error) {
	names, err := p.DefinitionStore.Names()
	if err != nil {
		return nil, err
	}
	definitions := []*parser.Definition{}
	for _, n := range names {
		if p.getServiceName(n) == name {
			data, err := p.DefinitionStore.Read(n)
			if err != nil {
				return nil, err
			}
			definition, err := p.parse(*data)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, definition)
		}
	}
	definition := p.merge(definitions)
	if definition != nil {
		p.applyPlugins(definition)
	}
	return definition, nil
}

func (p DefinitionProvider) merge(definitions []*parser.Definition) *parser.Definition {
	if len(definitions) == 0 {
		return nil
	}
	serviceName := p.getServiceName(definitions[0].Name)
	return MultiDefinition{}.Merge(serviceName, definitions)
}

func (p DefinitionProvider) getServiceName(name string) string {
	index := strings.Index(name, ".")
	if index != -1 {
		return name[:index]
	}
	return name
}

func (p DefinitionProvider) loadEmptyDefinitions() ([]DefinitionData, error) {
	names, err := p.DefinitionStore.Names()
	if err != nil {
		return nil, err
	}
	result := []DefinitionData{}
	for _, name := range names {
		serviceName := p.getServiceName(name)
		if len(result) == 0 || result[len(result)-1].Name != serviceName {
			result = append(result, *NewDefinitionData(serviceName, []byte{}))
		}
	}
	return result, nil
}

func (p DefinitionProvider) parse(data DefinitionData) (*parser.Definition, error) {
	definition, err := p.Parser.Parse(data.Name, data.Data)
	if err != nil {
		return nil, fmt.Errorf("Error parsing definition file '%s': %v", definition.Name, err)
	}
	return definition, nil
}

func (p DefinitionProvider) applyPlugins(definition *parser.Definition) {
	for _, plugin := range p.CommandPlugins {
		command := plugin.Command()
		if definition.Name == command.Service {
			p.applyPluginCommand(plugin, command, definition)
		}
	}
}

func (p DefinitionProvider) applyPluginCommand(plugin plugin.CommandPlugin, command plugin.Command, definition *parser.Definition) {
	parameters := p.convertToParameters(command.Parameters)
	var category *parser.OperationCategory
	if command.Category != nil {
		category = parser.NewOperationCategory(command.Category.Name, command.Category.Description)
	}
	baseUri, _ := url.Parse(parser.DefaultServerBaseUrl)
	operation := parser.NewOperation(command.Name, command.Description, "", *baseUri, "", "application/json", parameters, plugin, command.Hidden, category)
	for i := range definition.Operations {
		if definition.Operations[i].Name == command.Name {
			definition.Operations[i] = *operation
			return
		}
	}
	definition.Operations = append(definition.Operations, *operation)
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
