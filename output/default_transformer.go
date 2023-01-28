package output

type DefaultTransformer struct {
}

func (t DefaultTransformer) Execute(data interface{}) (interface{}, error) {
	return data, nil
}
