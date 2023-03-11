package commandline

import (
	"github.com/UiPath/uipathcli/parser"
)

// multiDefinition merges multiple definitions into a single one.
// This enables teams to provide fine-grained specifications for their individual
// micro-services and still provide them under a single command.
type multiDefinition struct{}

func (d multiDefinition) Merge(name string, definitions []*parser.Definition) *parser.Definition {
	if len(definitions) == 0 {
		return nil
	}
	operations := []parser.Operation{}
	for _, definition := range definitions {
		for _, operation := range definition.Operations {
			category := d.getCategory(operation, definition)
			operations = append(operations, *parser.NewOperation(operation.Name,
				operation.Summary,
				operation.Description,
				operation.Method,
				operation.BaseUri,
				operation.Route,
				operation.ContentType,
				operation.Parameters,
				operation.Plugin,
				operation.Hidden,
				category))
		}
	}
	return parser.NewDefinition(name, definitions[0].Description, operations)
}

func (d multiDefinition) getCategory(operation parser.Operation, definition *parser.Definition) *parser.OperationCategory {
	if operation.Category == nil || operation.Category.Description != "" {
		return operation.Category
	}
	return parser.NewOperationCategory(operation.Category.Name, definition.Description)
}

func newMultiDefinition() *multiDefinition {
	return &multiDefinition{}
}
