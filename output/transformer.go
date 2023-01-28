package output

type Transformer interface {
	Execute(data interface{}) (interface{}, error)
}
