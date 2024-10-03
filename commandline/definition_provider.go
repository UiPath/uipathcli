package commandline

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/plugin"
)

// The DefinitionProvider uses the store to read the definition files.
// It parses the definition files and also supports merging multiple files
// for the same service.
//
// The following mapping is performed between the files on disk and the commands
// the CLI provides:
// orchestrator.yaml => uipath orchestrator ...
// du.metering.yaml => uipath du ...
// du.events.yaml => uipath du ...
//
// For performance reasons, the definition provider always just loads the definition
// files belonging to a single service. There is no need to load the definition file
// for the du service when the user executes 'uipath orchestrator', for example.
type DefinitionProvider struct {
	store          DefinitionStore
	parser         parser.Parser
	commandPlugins []plugin.CommandPlugin
}

func (p DefinitionProvider) Index(serviceVersion string) ([]parser.Definition, error) {
	emptyDefinitions, err := p.loadEmptyDefinitions(serviceVersion)
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

func (p DefinitionProvider) Load(name string, serviceVersion string) (*parser.Definition, error) {
	names, err := p.store.Names(serviceVersion)
	if err != nil {
		return nil, err
	}
	definitions := []*parser.Definition{}
	for _, n := range names {
		if p.getServiceName(n) == name {
			data, err := p.store.Read(n, serviceVersion)
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
	return newMultiDefinition().Merge(serviceName, definitions)
}

func (p DefinitionProvider) getServiceName(name string) string {
	index := strings.Index(name, ".")
	if index != -1 {
		return name[:index]
	}
	return name
}

func (p DefinitionProvider) loadEmptyDefinitions(serviceVersion string) ([]DefinitionData, error) {
	names, err := p.store.Names(serviceVersion)
	if err != nil {
		return nil, err
	}
	result := []DefinitionData{}
	for _, name := range names {
		serviceName := p.getServiceName(name)
		if len(result) == 0 || result[len(result)-1].Name != serviceName {
			result = append(result, *NewDefinitionData(serviceName, serviceVersion, []byte{}))
		}
	}
	return result, nil
}

func (p DefinitionProvider) parse(data DefinitionData) (*parser.Definition, error) {
	definition, err := p.parser.Parse(data.Name, data.Data)
	if err != nil {
		return nil, fmt.Errorf("Error parsing definition file '%s': %w", data.Name, err)
	}
	return definition, nil
}

func (p DefinitionProvider) applyPlugins(definition *parser.Definition) {
	for _, plugin := range p.commandPlugins {
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
		category = parser.NewOperationCategory(command.Category.Name, command.Category.Summary, command.Category.Description)
	}
	baseUri, _ := url.Parse(parser.DefaultServerBaseUrl)
	operation := parser.NewOperation(command.Name, command.Description, "", "", *baseUri, "", "application/json", parameters, plugin, command.Hidden, category)
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
			parser.ParameterInCustom,
			p.Name,
			p.Required,
			nil,
			nil,
			[]parser.Parameter{})
		result = append(result, parameter)
	}
	return result
}

func NewDefinitionProvider(store DefinitionStore, parser parser.Parser, commandPlugins []plugin.CommandPlugin) *DefinitionProvider {
	return &DefinitionProvider{store, parser, commandPlugins}
}
