// Package plugin provides all the APIs and types for implementing custom commands.
//
// Plugins can provide commands which implement custom functionality. They can implement
// complex operations and provide convenient CLI commands for them.
package plugin

import (
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
)

// CommandPlugin is the interface plugin commands need to implement so
// they can be integrated with the CLI.
//
// The Command() operation defines the metadata for the command.
// The Execute() operation is invoked when the user runs the CLI command.
type CommandPlugin interface {
	Command() Command
	Execute(context ExecutionContext, writer output.OutputWriter, logger log.Logger) error
}
