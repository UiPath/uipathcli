package log

import (
	"bytes"
	"testing"
)

func TestLogErrorWritesToStandardError(t *testing.T) {
	var output bytes.Buffer
	logger := NewDebugLogger(&output)

	logger.LogError("There was an error")

	if output.String() != "There was an error" {
		t.Errorf("Standard error should contain error message, but got: %v", output.String())
	}
}

func TestLogRequestDisplaysRequestDetails(t *testing.T) {
	var output bytes.Buffer
	logger := NewDebugLogger(&output)

	body := bytes.NewBufferString(`{"hello":"world"}`)
	header := map[string][]string{
		"x-request-id": {"my-request-id"},
	}
	logger.LogRequest(*NewRequestInfo("POST", "https://cloud.uipath.com/my-service", "HTTP/1.1", header, body))

	expectedOutput := `POST https://cloud.uipath.com/my-service HTTP/1.1
x-request-id: my-request-id

{"hello":"world"}


`
	if output.String() != expectedOutput {
		t.Errorf("Standard output should contain request, but got: %v", output.String())
	}
}

func TestLogResponseDisplaysResponseDetails(t *testing.T) {
	var output bytes.Buffer
	logger := NewDebugLogger(&output)

	body := bytes.NewBufferString(`{"hello":"world"}`)
	header := map[string][]string{
		"x-request-id": {"my-request-id"},
	}
	logger.LogResponse(*NewResponseInfo(200, "200 OK", "HTTP/1.1", header, body))

	expectedOutput := `HTTP/1.1 200 OK
x-request-id: my-request-id

{"hello":"world"}


`
	if output.String() != expectedOutput {
		t.Errorf("Standard output should contain request, but got: %v", output.String())
	}
}
