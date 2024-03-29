package output

import (
	"fmt"

	"github.com/jmespath/go-jmespath"
)

// The JmesPathTransformer uses the JMESPath query language to transform the executor output.
//
// It is used when the --query parameter is provided.
// Example: --query "id"
//
//	{
//	  "id": "521b4edc-ad6f-4301-909e-f96a401e1fed",
//	}
//
// => "521b4edc-ad6f-4301-909e-f96a401e1fed"
//
// See https://jmespath.org for more information.
type JmesPathTransformer struct {
	query string
}

func (t JmesPathTransformer) Execute(data interface{}) (interface{}, error) {
	result, err := jmespath.Search(t.query, data)
	if err != nil {
		return nil, fmt.Errorf("Error in query: %w", err)
	}
	return result, nil
}

func NewJmesPathTransformer(query string) *JmesPathTransformer {
	return &JmesPathTransformer{query}
}
