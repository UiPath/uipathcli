package output

import (
	"bytes"
	"testing"
)

func TestJsonWriterOutputsErrorStatusWhenResponseIsFailure(t *testing.T) {
	var output bytes.Buffer
	writer := NewJsonOutputWriter(&output, NewDefaultTransformer())

	err := writer.WriteResponse(*NewResponseInfo(400, "400 BadRequest", "HTTP/1.1", map[string][]string{}, bytes.NewReader([]byte{})))

	if err != nil {
		t.Errorf("Writing response failed: %v", err)
	}
	if output.String() != "HTTP/1.1 400 BadRequest\n" {
		t.Errorf("Should show HTTP error status, but got: %v", output.String())
	}
}

func TestJsonWriterOutputsResponseBody(t *testing.T) {
	output := bytes.NewBufferString(`{"hello":"world"}`)
	writer := NewJsonOutputWriter(output, NewDefaultTransformer())

	err := writer.WriteResponse(*NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, output))

	if err != nil {
		t.Errorf("Writing response failed: %v", err)
	}
	if output.String() != `{
  "hello": "world"
}
` {
		t.Errorf("Should show response value, but got: %v", output.String())
	}
}

func TestJsonWriterOutputsPlainBodyOnJsonParsingError(t *testing.T) {
	output := bytes.NewBufferString(`{invalid}`)
	writer := NewJsonOutputWriter(output, NewDefaultTransformer())

	err := writer.WriteResponse(*NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, output))

	if err != nil {
		t.Errorf("Writing response failed: %v", err)
	}
	if output.String() != "{invalid}" {
		t.Errorf("Should show response plain body, but got: %v", output.String())
	}
}
