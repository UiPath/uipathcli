package parser

import "net/url"

type Definition struct {
	Name        string
	BaseUri     url.URL
	Description string
	Operations  []Operation
}

func NewDefinition(name string, baseUri url.URL, description string, operations []Operation) *Definition {
	return &Definition{name, baseUri, description, operations}
}
