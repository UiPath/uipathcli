// Package parser parses OpenAPI specifications, extracts all the information needed
// to create the CLI commands and parameters.
package parser

// The Parser interface provides an abstraction for parsing the service definition
// files. It returns a structured Definition document with all the operations and
// parameters of the service.
type Parser interface {
	Parse(name string, data []byte) (*Definition, error)
}
