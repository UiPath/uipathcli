package plugin

// MultiCommandPlugin can be implemented to provide dynamic list of commands
// which are all handled by a single command plugin.
//
// The Commands() operation defines the metadata for multiple commands.
type MultiCommandPlugin interface {
	Commands() []Command
}
