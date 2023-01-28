package output

import (
	"fmt"

	"github.com/jmespath/go-jmespath"
)

type JmesPathTransformer struct {
	Query string
}

func (t JmesPathTransformer) Execute(data interface{}) (interface{}, error) {
	result, err := jmespath.Search(t.Query, data)
	if err != nil {
		return nil, fmt.Errorf("Error in query: %v", err)
	}
	return result, nil
}
