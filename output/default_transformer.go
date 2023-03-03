package output

// The DefaultTransformer simply passes through the output from the output writer.
// No transformation is performed.
type DefaultTransformer struct {
}

func (t DefaultTransformer) Execute(data interface{}) (interface{}, error) {
	return data, nil
}
