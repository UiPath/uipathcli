package output

// The Transformer interface can be implemented to provide a converter which transforms
// the CLI output into a different structure.
type Transformer interface {
	Execute(data interface{}) (interface{}, error)
}
