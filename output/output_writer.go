// Package output formats and prints the response on standard output.
//
// - Supports JSON and text output
// - Provides mechanism to transform output using JMESPath queries
package output

// The OutputWriter is used to print the executor result on standard output.
type OutputWriter interface {
	WriteResponse(response ResponseInfo) error
}
