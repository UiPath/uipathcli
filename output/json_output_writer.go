package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// The JsonOutputWriter formats the CLI output as prettified json.
//
// It is used by default or when the --output json parameter is provided.
// Example:
//
//	{
//	 "foo": "bar"
//	}
type JsonOutputWriter struct {
	Output      io.Writer
	Transformer Transformer
}

func (w JsonOutputWriter) writeBody(body []byte) error {
	var data interface{}
	err := json.Unmarshal(body, &data)
	if err != nil {
		fmt.Fprint(w.Output, string(body))
		return nil
	}

	transformedResult, err := w.Transformer.Execute(data)
	if err != nil {
		return err
	}
	result, err := json.MarshalIndent(transformedResult, "", "  ")
	if err != nil {
		return err
	}
	w.Output.Write(result)
	w.Output.Write([]byte("\n"))
	return nil
}

func (w JsonOutputWriter) WriteResponse(response ResponseInfo) error {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 && response.StatusCode >= 400 {
		fmt.Fprintf(w.Output, "%s %s\n", response.Protocol, response.Status)
		return nil
	}
	return w.writeBody(body)
}
