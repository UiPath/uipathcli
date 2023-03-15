package output

import (
	"bytes"
	"testing"
)

func TestTextWriterOutputsErrorStatusWhenResponseIsFailure(t *testing.T) {
	var output bytes.Buffer
	writer := NewTextOutputWriter(&output, NewDefaultTransformer())

	err := writer.WriteResponse(*NewResponseInfo(400, "400 BadRequest", "HTTP/1.1", map[string][]string{}, bytes.NewReader([]byte{})))

	if err != nil {
		t.Errorf("Writing response failed: %v", err)
	}
	if output.String() != "HTTP/1.1 400 BadRequest\n" {
		t.Errorf("Should show HTTP error status, but got: %v", output.String())
	}
}

func TestTextWriterOutputsResponseBody(t *testing.T) {
	output := bytes.NewBufferString(`{"hello":"world"}`)
	writer := NewTextOutputWriter(output, NewDefaultTransformer())

	err := writer.WriteResponse(*NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, output))

	if err != nil {
		t.Errorf("Writing response failed: %v", err)
	}
	if output.String() != "world\n" {
		t.Errorf("Should show response value, but got: %v", output.String())
	}
}

func TestTextWriterOutputsResponseBodySortedByKeys(t *testing.T) {
	output := bytes.NewBufferString(`{"b":"world","a":"hello"}`)
	writer := NewTextOutputWriter(output, NewDefaultTransformer())

	err := writer.WriteResponse(*NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, output))

	if err != nil {
		t.Errorf("Writing response failed: %v", err)
	}
	if output.String() != "hello\tworld\n" {
		t.Errorf("Should show response body sorted by keys, but got: %v", output.String())
	}
}

func TestTextWriterOutputsResponseBodyArray(t *testing.T) {
	output := bytes.NewBufferString(`["hello","world"]`)
	writer := NewTextOutputWriter(output, NewDefaultTransformer())

	err := writer.WriteResponse(*NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, output))

	if err != nil {
		t.Errorf("Writing response failed: %v", err)
	}
	if output.String() != "hello\nworld\n" {
		t.Errorf("Should show response body array, but got: %v", output.String())
	}
}

func TestTextWriterOutputsResponseBodyObjectArray(t *testing.T) {
	output := bytes.NewBufferString(`[{"a":"hello"},{"a":"world"}]`)
	writer := NewTextOutputWriter(output, NewDefaultTransformer())

	err := writer.WriteResponse(*NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, output))

	if err != nil {
		t.Errorf("Writing response failed: %v", err)
	}
	if output.String() != "hello\nworld\n" {
		t.Errorf("Should show response body array, but got: %v", output.String())
	}
}

func TestTextWriterOutputsResponseBodyObjectArrayDifferentKeys(t *testing.T) {
	output := bytes.NewBufferString(`[{"b":"foo","a":"hello"},{"b":"bar"}]`)
	writer := NewTextOutputWriter(output, NewDefaultTransformer())

	err := writer.WriteResponse(*NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, output))

	if err != nil {
		t.Errorf("Writing response failed: %v", err)
	}
	if output.String() != "hello\tfoo\n\tbar\n" {
		t.Errorf("Should show response body object array, but got: %v", output.String())
	}
}

func TestTextWriterOutputsPlainBodyOnJsonParsingError(t *testing.T) {
	output := bytes.NewBufferString(`{invalid}`)
	writer := NewTextOutputWriter(output, NewDefaultTransformer())

	err := writer.WriteResponse(*NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, output))

	if err != nil {
		t.Errorf("Writing response failed: %v", err)
	}
	if output.String() != "{invalid}" {
		t.Errorf("Should show plain body, but got: %v", output.String())
	}
}
