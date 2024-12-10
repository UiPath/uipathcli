// Package log provides an API to log diagnostics information which is used
// to give the user more details about what is happening inside of the CLI.
//
// - Stores full request and response information
// - Provides data to root-cause problems
package log

// The Logger interface which is used to provide additional information to the
// user about what operations the CLI is performing.
type Logger interface {
	Log(message string)
	LogError(message string)
	LogRequest(request RequestInfo)
	LogResponse(response ResponseInfo)
}
