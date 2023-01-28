package output

import (
	"encoding/json"
	"fmt"
	"io"
)

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
	if len(response.Body) == 0 && response.StatusCode >= 400 {
		fmt.Fprintf(w.Output, "%s %s\n", response.Protocol, response.Status)
		return nil
	}
	return w.writeBody(response.Body)
}
